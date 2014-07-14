package sum

import (
	"archive/tar"
	"bytes"
	"io"
	"io/ioutil"
	"path"

	"github.com/dotcloud/docker/pkg/tarsum"
)

// .. this is an all-in-one. I wish this could be an iterator.
func SumAllDockerSave(saved io.Reader) (map[string]string, error) {
	tarRdr := tar.NewReader(saved)
	hashes := map[string]string{}
	jsons := map[string][]byte{}
	for {
		hdr, err := tarRdr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if path.Base(hdr.Name) == "json" {
			id := path.Dir(hdr.Name)
			jsonBuf, err := ioutil.ReadAll(tarRdr)
			if err != nil {
				if err == io.EOF {
					continue
				}
				return nil, err
			}
			jsons[id] = jsonBuf
		}
		if path.Base(hdr.Name) == "layer.tar" {
			id := path.Dir(hdr.Name)
			jsonRdr := bytes.NewReader(jsons[id])
			delete(jsons, id)
			sum, err := SumTarLayer(tarRdr, jsonRdr, nil)
			if err != nil {
				if err == io.EOF {
					continue
				}
				return nil, err
			}
			hashes[id] = sum
		}
	}
	return hashes, nil
}

// if out is not nil, then the tar input is written there instead
func SumTarLayer(tarReader io.Reader, json io.Reader, out io.Writer) (string, error) {
	var writer io.Writer = ioutil.Discard
	if out != nil {
		writer = out
	}
	ts := &tarsum.TarSum{Reader: tarReader}
	_, err := io.Copy(writer, ts)
	if err != nil {
		return "", err
	}

	buf, err := ioutil.ReadAll(json)
	if err != nil {
		return "", err
	}

	return ts.Sum(buf), nil
}
