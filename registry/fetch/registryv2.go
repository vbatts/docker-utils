package fetch

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

var defaultV2Routes = mux.NewRouter()

func init() {
	defaultV2Routes.Path("/v2/").Name("Base")
	defaultV2Routes.Path("/v2/{name:.*}/tags/list").Name("ImageTagList")
	defaultV2Routes.Path("/v2/{name:.*}/manifests/{reference}").Name("ImageManifest")
	defaultV2Routes.Path("/v2/{name:.*}/blobs/{blob}").Name("ImageBlob")
}

type registryV2Endpoint struct {
	scheme       string
	schemeTested bool
	host         string
	tokens       map[string]Token
	endpoints    []string
}

func (re *registryV2Endpoint) Pull(img ImageRef, dest string) error {
	return nil
}

// stub to satisfy the interface
func (re *registryV2Endpoint) ImageID(img ImageRef) (string, error) {
	return img.ID(), nil
}

// two things: check whether this is a v2 registry, and determine https:// or http://
func (re *registryV2Endpoint) ping() error {
	url, err := defaultV2Routes.Get("Base").URL()
	if err != nil {
		return err
	}
	if re.schemeTested {
		url.Scheme = re.scheme
	} else {
		url.Scheme = "https" // try first
	}
	url.Host = re.host
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		// ugly Failback to http
		if strings.Contains(err.Error(), "tls: oversized record received with length") {
			url.Scheme = "http"
			req, err = http.NewRequest("GET", url.String(), nil)
			if err != nil {
				return err
			}
			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer resp.Body.Close()

	re.schemeTested = true
	re.scheme = url.Scheme

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Get(%q) returned %q", url, resp.Status)
	}
	return nil
}

// Token fetches and returns a fresh Token from this registryV1Endpoint for the imageName provided
func (re *registryV2Endpoint) Token(img ImageRef) (Token, error) {
	// TODO will need to take the 'Www-Authenticate' header to collect a token.
	// For the docker hub this may likely go to https://auth.docker.io/token

	return emptyToken, nil
}

// FIXME this will need to be under registryV2Endpoint so it can reuse scheme and host and endpoints
func getV2TagList(img ImageRef) (*v2TagList, error) {
	url, err := defaultV2Routes.Get("ImageTagList").URL("name", img.Name())
	if err != nil {
		return nil, err
	}
	url.Scheme = "https" // try first
	url.Host = img.Host()
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "tls: oversized record received with length") {
			// ugly Failback to http
			url.Scheme = "http"
			req, err = http.NewRequest("GET", url.String(), nil)
			if err != nil {
				return nil, err
			}
			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Get(%q) returned %q", url, resp.Status)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	tl := v2TagList{}
	if err := json.Unmarshal(buf, &tl); err != nil {
		return nil, err
	}

	return &tl, nil
}

// {"name":"vbatts/slackware","tags":["latest"]}
type v2TagList struct {
	Name string
	Tags []string
}

// caller needs to close the io.ReadCloser
func getV2Blob(img ImageRef, blob string) (int64, io.ReadCloser, error) {
	url, err := defaultV2Routes.Get("ImageBlob").URL("name", img.Name(), "blob", blob)
	if err != nil {
		return 0, nil, err
	}
	url.Scheme = "https" // try first
	url.Host = img.Host()
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return 0, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "tls: oversized record received with length") {
			// ugly Failback to http
			url.Scheme = "http"
			req, err = http.NewRequest("GET", url.String(), nil)
			if err != nil {
				return 0, nil, err
			}
			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return 0, nil, err
			}
		} else {
			return 0, nil, err
		}
	}

	if resp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("Get(%q) returned %q", url, resp.Status)
	}

	return resp.ContentLength, resp.Body, nil
}

func getV2Manifest(img ImageRef) (*v2Manifest, error) {
	url, err := defaultV2Routes.Get("ImageManifest").URL("name", img.Name(), "reference", img.Tag())
	if err != nil {
		return nil, err
	}
	url.Scheme = "https" // try first
	url.Host = img.Host()
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "tls: oversized record received with length") {
			// ugly Failback to http
			url.Scheme = "http"
			req, err = http.NewRequest("GET", url.String(), nil)
			if err != nil {
				return nil, err
			}
			resp, err = http.DefaultClient.Do(req)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Get(%q) returned %q", url, resp.Status)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := v2Manifest{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

type v2Manifest struct {
	SchemaVersion int `json:"schemaVersion"`
	Name          string
	Tag           string
	Architecture  string
	FSLayers      []map[string]string `json:"fsLayers"`
	History       []map[string]string `json:"history"`
	str           string
}
