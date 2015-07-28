package registry

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestImageRefHost(t *testing.T) {
	/*
					   - docker.io/tianon/true
					   - docker.io/golang
				     - tianon/true
		         - fedora
		         - localhost:5000/fedora
		         - 192.168.1.23:5000/fedora
		         - 192.168.1.23/fedora
	*/
	cases := []struct {
		Name     string
		Expected string
	}{
		{"docker.io/tianon/true", DefaultHubNamespace},
		{"docker.io:80/tianon/true", DefaultHubNamespace + ":80"},
		{"docker.io/tianon/true:latest", DefaultHubNamespace},
		{"tianon/true", DefaultHubNamespace},
		{"tianon/true:latest", DefaultHubNamespace},
		{"fedora:latest", DefaultHubNamespace},
		{"localhost:5000/fedora", "localhost:5000"},
		{"localhost:5000/fedora:latest", "localhost:5000"},
		{"localhost/fedora", "localhost"},
		{"localhost/fedora:latest", "localhost"},
		{"192.168.1.23:5000/tianon/true", "192.168.1.23:5000"},
		{"192.168.1.23:5000/fedora", "192.168.1.23:5000"},
		{"192.168.1.23/fedora", "192.168.1.23"},
		{"192.168.1.23/fedora:latest", "192.168.1.23"},
		{"192.168.1.23/library/fedora", "192.168.1.23"},
	}
	for _, c := range cases {
		ref := NewImageRef(c.Name)
		if ref.Host() != c.Expected {
			t.Errorf("from %q: expected %q, got %q", c.Name, c.Expected, ref.Host())
		}
	}
}

func TestRegistryFetchToken(t *testing.T) {
	ref := NewImageRef("tianon/true")
	r := NewRegistry(ref.Host())
	tok, err := r.Token(ref)
	if err != nil {
		t.Fatal(err)
	}
	if tok.Signature() == "" {
		t.Errorf("expected Signature, but it was empty")
	}
	if tok.Repository() == "" {
		t.Errorf("expected Repository, but it was empty")
	}
	if tok.Access() == "" {
		t.Errorf("expected Access, but it was empty")
	}
}
func TestRegistryFetchImageID(t *testing.T) {
	r := NewRegistry(DefaultRegistryHost)
	id, err := r.ImageID(NewImageRef("tianon/true"))
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Errorf("expected an ImageID, but it was empty")
	}
}
func TestRegistryFetchAncestry(t *testing.T) {
	r := NewRegistry(DefaultRegistryHost)
	ids, err := r.Ancestry(NewImageRef("tianon/true"))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) == 0 {
		t.Errorf("expected an Ancestry, but it was empty")
	}
}
func TestRegistryFetchLayers(t *testing.T) {
	r := NewRegistry(DefaultRegistryHost)
	tdir, err := ioutil.TempDir("", "test.fetch.")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tdir)

	layersFetched, err := r.FetchLayers(NewImageRef("tianon/true"), tdir)
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range layersFetched {
		if _, err := os.Stat(path.Join(tdir, id, "json")); err != nil {
			t.Error(err)
		}
	}
}
func TestRegistryImageRepositoriesFile(t *testing.T) {
	r := NewRegistry(DefaultRegistryHost)
	buf, err := r.FormatRepositories(NewImageRef("tianon/true"))
	if err != nil {
		t.Fatal(err)
	}
	if len(buf) == 0 {
		t.Errorf("expected a populated `repositories` info")
	}
}
