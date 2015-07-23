package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

var DefaultRegistryHost = "index.docker.io"

func NewImageRef(name string) *ImageRef {
	return &ImageRef{orig: name}
}

type ImageRef struct {
	orig     string
	name     string
	tag      string
	id       string
	ancestry []string
}

func (ir ImageRef) Host() string {
	return ""
}

func (ir ImageRef) ID() string {
	return ir.id
}
func (ir *ImageRef) SetID(id string) {
	ir.id = id
}

func (ir ImageRef) Ancestry() []string {
	return ir.ancestry
}
func (ir *ImageRef) SetAncestry(ids []string) {
	ir.ancestry = make([]string, len(ids))
	for i := range ids {
		ir.ancestry[i] = ids[i]
	}
}
func (ir ImageRef) Name() string {
	count := strings.Count(ir.orig, ":")
	if count == 0 {
		return ir.orig
	}
	if count == 1 {
		return strings.Split(ir.orig, ":")[0]
	}
	return ""
}
func (ir ImageRef) Tag() string {
	if ir.tag != "" {
		return ir.tag
	}
	count := strings.Count(ir.orig, ":")
	if count == 0 {
		return "latest"
	}
	if count == 1 {
		return strings.Split(ir.orig, ":")[1]
	}
	return ""
}
func (ir ImageRef) Digest() string {
	return ""
}

func (ir ImageRef) String() string {
	return ir.orig
}

func NewRegistry(host string) RegistryEndpoint {
	return RegistryEndpoint{
		Host:      host,
		tokens:    map[string]Token{},
		endpoints: []string{},
	}
}

type RegistryEndpoint struct {
	Host      string
	tokens    map[string]Token
	endpoints []string
}

// Token fetches and returns a fresh Token from this RegistryEndpoint for the imageName provided
func (re *RegistryEndpoint) Token(img *ImageRef) (Token, error) {
	url := fmt.Sprintf("https://%s/v1/repositories/%s/images", re.Host, img.Name())
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

	//logrus.Infof("%#v", resp)

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

func (re *RegistryEndpoint) ImageID(img *ImageRef) (string, error) {
	if _, ok := re.tokens[img.Name()]; !ok {
		if _, err := re.Token(img); err != nil {
			return "", err
		}
	}
	endpoint := re.Host
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

	//logrus.Infof("%#v", resp)
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	str := strings.Trim(string(buf), "\"")
	img.SetID(str)
	return img.ID(), nil
}

func (re *RegistryEndpoint) Ancestry(img *ImageRef) ([]string, error) {
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

	endpoint := re.Host
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

	//logrus.Infof("%#v", resp)
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

func (re *RegistryEndpoint) FetchLayers(img *ImageRef, dest string) ([]string, error) {
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

	endpoint := re.Host
	if len(re.endpoints) > 0 {
		endpoint = re.endpoints[0]
	}
	for _, id := range img.Ancestry() {
		if err := os.MkdirAll(path.Join(dest, id), 0755); err != nil {
			return emptySet, err
		}
		// get the json file first
		err := func() error {
			url := fmt.Sprintf("https://%s/v1/images/%s/json", endpoint, img.ID())
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

			//logrus.Infof("%#v", resp)
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
		/*
			err := func() err {
				url := fmt.Sprintf("https://%s/v1/images/%s/json", endpoint, img.ID())
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

				//logrus.Infof("%#v", resp)
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
		*/
	}

	return img.Ancestry(), nil
}

var (
	// ErrTokenHeaderEmpty if the response from the registry did not provide a Token
	ErrTokenHeaderEmpty = fmt.Errorf("HTTP Header x-docker-token is empty")

	emptyToken = Token("")
)

// Token is access token from a docker registry
type Token string

func (t Token) Signature() string {
	return t.getFieldValue("Signature")
}

func (t Token) Repository() string {
	return t.getFieldValue("Repository")
}

func (t Token) Access() string {
	return t.getFieldValue("Access")
}

func (t Token) getFieldValue(key string) string {
	for _, part := range strings.Split(t.String(), ",") {
		if strings.HasPrefix(strings.ToLower(part), strings.ToLower(key)) {
			chunks := strings.SplitN(part, "=", 2)
			if len(chunks) > 2 {
				continue
			}
			return chunks[1]
		}
	}
	return ""
}

// String to satisfy the fmt.Stringer interface
func (t Token) String() string {
	return string(t)
}
