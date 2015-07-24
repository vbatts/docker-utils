package registry

import (
	"io/ioutil"
	"testing"
)

func TestRegistryFetchToken(t *testing.T) {
	r := NewRegistry(DefaultRegistryHost)
	tok, err := r.Token(NewImageRef("vbatts/slackware"))
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
	id, err := r.ImageID(NewImageRef("vbatts/slackware"))
	if err != nil {
		t.Fatal(err)
	}
	if id == "" {
		t.Errorf("expected an ImageID, but it was empty")
	}
}
func TestRegistryFetchAncestry(t *testing.T) {
	r := NewRegistry(DefaultRegistryHost)
	ids, err := r.Ancestry(NewImageRef("vbatts/slackware"))
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
	_, err = r.FetchLayers(NewImageRef("vbatts/slackware"), tdir)
	if err != nil {
		t.Fatal(err)
	}

	t.Fatal(tdir)
}
