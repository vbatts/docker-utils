package registry

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestRegistryFetchToken(t *testing.T) {
	r := NewRegistry(DefaultRegistryHost)
	tok, err := r.Token(NewImageRef("tianon/true"))
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
