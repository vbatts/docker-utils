package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Sirupsen/logrus"
)

type registryV1Endpoint struct {
	host      string
	tokens    map[string]Token
	endpoints []string
}

func (re *registryV1Endpoint) Host() string {
	return re.host
}

// Token fetches and returns a fresh Token from this registryV1Endpoint for the imageName provided
func (re *registryV1Endpoint) Token(img ImageRef) (Token, error) {
	url := fmt.Sprintf("https://%s/v1/repositories/%s/images", re.host, img.Name())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return emptyToken, err
	}
	req.Header.Add("X-Docker-Token", "true")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return emptyToken, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return emptyToken, fmt.Errorf("Get(%q) returned %q", url, resp.Status)
	}

	//logrus.Debugf("%#v", resp)

	// looking for header: X-Docker-Token: signature=4709c3e8d96f6a0e9fa53bd205b5be171ac9ade0,repository="vbatts/slackware",access=read
	tok := resp.Header.Get("X-Docker-Token")
	if tok == "" {
		return emptyToken, ErrTokenHeaderEmpty
	}
	endpoint := resp.Header.Get("X-Docker-Endpoints")
	if endpoint != "" {
		re.endpoints = append(re.endpoints, endpoint)
	}

	re.tokens[img.Name()] = Token(tok)
	return re.tokens[img.Name()], nil
}

func (re *registryV1Endpoint) ImageID(img ImageRef) (string, error) {
	if _, ok := re.tokens[img.Name()]; !ok {
		if _, err := re.Token(img); err != nil {
			return "", err
		}
	}
	endpoint := re.host
	if len(re.endpoints) > 0 {
		endpoint = re.endpoints[0]
	}
	url := fmt.Sprintf("https://%s/v1/repositories/%s/tags/%s", endpoint, img.Name(), img.Tag())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", re.tokens[img.Name()]))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Get(%q) returned %q", url, resp.Status)
	}

	//logrus.Debugf("%#v", resp)
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	str := strings.Trim(string(buf), "\"")
	img.SetID(str)
	return img.ID(), nil
}

func (re *registryV1Endpoint) Ancestry(img ImageRef) ([]string, error) {
	emptySet := []string{}
	if _, ok := re.tokens[img.Name()]; !ok {
		if _, err := re.Token(img); err != nil {
			return emptySet, err
		}
	}
	if img.ID() == "" {
		if _, err := re.ImageID(img); err != nil {
			return emptySet, err
		}
	}

	endpoint := re.host
	if len(re.endpoints) > 0 {
		endpoint = re.endpoints[0]
	}
	url := fmt.Sprintf("https://%s/v1/images/%s/ancestry", endpoint, img.ID())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return emptySet, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", re.tokens[img.Name()]))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return emptySet, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return emptySet, fmt.Errorf("Get(%q) returned %q", url, resp.Status)
	}

	//logrus.Debugf("%#v", resp)
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return emptySet, err
	}

	set := []string{}
	if err := json.Unmarshal(buf, &set); err != nil {
		return emptySet, err
	}
	img.SetAncestry(set)
	return img.Ancestry(), nil
}

// This is presently fetching docker-registry v1 API and returns the IDs of the layers fetched from the registry
func (re *registryV1Endpoint) FetchLayers(img ImageRef, dest string) ([]string, error) {
	emptySet := []string{}
	if _, ok := re.tokens[img.Name()]; !ok {
		if _, err := re.Token(img); err != nil {
			return emptySet, err
		}
	}
	if img.ID() == "" {
		if _, err := re.ImageID(img); err != nil {
			return emptySet, err
		}
	}
	if len(img.Ancestry()) == 0 {
		if _, err := re.Ancestry(img); err != nil {
			return emptySet, err
		}
	}

	endpoint := re.host
	if len(re.endpoints) > 0 {
		endpoint = re.endpoints[0]
	}
	for _, id := range img.Ancestry() {
		logrus.Debugf("Fetching layer %s", id)
		if err := os.MkdirAll(path.Join(dest, id), 0755); err != nil {
			return emptySet, err
		}
		// get the json file first
		err := func() error {
			url := fmt.Sprintf("https://%s/v1/images/%s/json", endpoint, id)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}
			req.Header.Add("Authorization", fmt.Sprintf("Token %s", re.tokens[img.Name()]))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("Get(%q) returned %q", url, resp.Status)
			}

			//logrus.Debugf("%#v", resp)
			fh, err := os.Create(path.Join(dest, id, "json"))
			if err != nil {
				return err
			}
			defer fh.Close()
			if _, err := io.Copy(fh, resp.Body); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return emptySet, err
		}

		// get the layer file next
		err = func() error {
			url := fmt.Sprintf("https://%s/v1/images/%s/layer", endpoint, id)
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}
			logrus.Debugf("%q", fmt.Sprintf("Token %s", re.tokens[img.Name()]))
			req.Header.Add("Authorization", fmt.Sprintf("Token %s", re.tokens[img.Name()]))

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("Get(%q) returned %q", url, resp.Status)
			}

			logrus.Debugf("[FetchLayers] ended up at %q", resp.Request.URL.String())
			logrus.Debugf("[FetchLayers] response %#v", resp)
			fh, err := os.Create(path.Join(dest, id, "layer.tar"))
			if err != nil {
				return err
			}
			defer fh.Close()
			if _, err := io.Copy(fh, resp.Body); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return emptySet, err
		}
	}

	return img.Ancestry(), nil
}
