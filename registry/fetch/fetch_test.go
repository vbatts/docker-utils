package registry

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestImageRefHost(t *testing.T) {
	cases := []struct {
		Name         string
		ExpectedHost string
		ExpectedName string
		ExpectedTag  string
	}{
		{"docker.io/tianon/true", DefaultHubNamespace, "tianon/true", DefaultTag},
		{"docker.io:80/tianon/true", DefaultHubNamespace + ":80", "tianon/true", DefaultTag},
		{"docker.io/tianon/true:hurr", DefaultHubNamespace, "tianon/true", "hurr"},
		{"tianon/true", DefaultHubNamespace, "tianon/true", DefaultTag},
		{"tianon/true:latest", DefaultHubNamespace, "tianon/true", DefaultTag},
		{"fedora:latest", DefaultHubNamespace, "fedora", DefaultTag},
		{"localhost:5000/fedora", "localhost:5000", "fedora", DefaultTag},
		{"localhost:5000/fedora:latest", "localhost:5000", "fedora", DefaultTag},
		{"localhost/fedora", "localhost", "fedora", DefaultTag},
		{"localhost/fedora:latest", "localhost", "fedora", DefaultTag},
		{"192.168.1.23:5000/tianon/true", "192.168.1.23:5000", "tianon/true", DefaultTag},
		{"192.168.1.23:5000/fedora", "192.168.1.23:5000", "fedora", DefaultTag},
		{"192.168.1.23/fedora", "192.168.1.23", "fedora", DefaultTag},
		{"192.168.1.23/fedora:latest", "192.168.1.23", "fedora", DefaultTag},
		{"192.168.1.23/library/fedora", "192.168.1.23", "library/fedora", DefaultTag},
	}
	for _, c := range cases {
		ref := NewImageRef(c.Name)
		if ref.Host() != c.ExpectedHost {
			t.Errorf("from %q: expected %q, got %q", c.Name, c.ExpectedHost, ref.Host())
		}
		if ref.Name() != c.ExpectedName {
			t.Errorf("from %q: expected %q, got %q", c.Name, c.ExpectedName, ref.Name())
		}
		if ref.Tag() != c.ExpectedTag {
			t.Errorf("from %q: expected %q, got %q", c.Name, c.ExpectedTag, ref.Tag())
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
