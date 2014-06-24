package sum

import (
	"archive/tar"
	"bytes"
	"github.com/dotcloud/docker/utils"
	"io"
	"io/ioutil"
	"path"
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
			sum, err := SumTarLayer(tarRdr, jsonRdr)
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

func SumTarLayer(tarReader io.Reader, json io.Reader) (string, error) {
	ts := &utils.TarSum{Reader: tarReader}
	_, err := io.Copy(ioutil.Discard, ts)
	if err != nil {
		return "", err
	}

	buf, err := ioutil.ReadAll(json)
	if err != nil {
		return "", err
	}

	return ts.Sum(buf), nil
}
