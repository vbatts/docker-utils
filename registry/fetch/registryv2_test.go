package fetch

import (
	"crypto/sha256"
	"fmt"
	"io"
	"testing"
)

func TestV2Ping(t *testing.T) {
	re := registryV2Endpoint{host: "localhost:5000"}
	re.ping()
}

func TestUrl(t *testing.T) {
	img := NewImageRef("localhost:5000/vbatts/slackware")
	tl, err := getV2TagList(img)
	if err != nil {
		t.Fatalf("%#v %t", err, err)
	}
	img.SetTag(tl.Tags[0])
	manifest, err := getV2Manifest(img)
	if err != nil {
		t.Fatalf("%#v %t", err, err)
	}

	for _, bs := range manifest.FSLayers {
		h := sha256.New()
		i, rdr, err := getV2Blob(img, bs["blobSum"])
		if err != nil {
			t.Fatalf("%#v %t", err, err)
		}
		j, err := io.Copy(h, rdr)
		if err != nil {
			t.Fatal(err)
		}
		if i != j {
			t.Errorf("expected %d length, got %d", i, j)
		}
		if bs["blobSum"] != fmt.Sprintf("sha256:%x", h.Sum(nil)) {
			t.Errorf("expected %q sum, got %q", bs["blobSum"], fmt.Sprintf("sha256:%x", h.Sum(nil)))
		}
	}
}
