package fetch

import (
	"encoding/json"
	"fmt"
	"strings"
)

var (
	DefaultRegistryHost = "index.docker.io"
	DefaultHubNamespace = "docker.io"
	DefaultTag          = "latest"
)

type RegistryEndpoint interface {
	Host() string
	Token(*ImageRef) (Token, error)
	ImageID(*ImageRef) (string, error)
	Ancestry(*ImageRef) ([]string, error)
	FetchLayers(*ImageRef, string) ([]string, error)
}

func NewRegistry(host string) RegistryEndpoint {
	if host == "docker.io" {
		host = DefaultRegistryHost
	}
	return &registryV1Endpoint{
		host:      host,
		tokens:    map[string]Token{},
		endpoints: []string{},
	}
}

// Return the `repositories` file format data for the referenced image
func FormatRepositories(refs ...*ImageRef) ([]byte, error) {
	// new Registry, ref.host function
	for _, ref := range refs {
		if ref.ID() == "" {
			re := NewRegistry(ref.Host())
			if _, err := re.ImageID(ref); err != nil {
				return nil, err
			}
		}
	}
	// {"busybox":{"latest":"4986bf8c15363d1c5d15512d5266f8777bfba4974ac56e3270e7760f6f0a8125"}}
	repoInfo := map[string]map[string]string{}
	for _, ref := range refs {
		if repoInfo[ref.Name()] == nil {
			repoInfo[ref.Name()] = map[string]string{ref.Tag(): ref.ID()}
		} else {
			repoInfo[ref.Name()][ref.Tag()] = ref.ID()
		}
	}
	return json.Marshal(repoInfo)
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
