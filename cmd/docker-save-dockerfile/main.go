package main

import (
	"archive/tar"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/vbatts/docker-utils/dockerfile"
)

var (
	flVerbose = flag.Bool("v", false, "turn on verbose debug")
)

func main() {
	flag.Parse()
	if *flVerbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	var tarInput io.Reader
	if flag.NArg() == 0 {
		tarInput = os.Stdin
		logrus.Info("using stdin ...")
	} else {
		fh, err := os.Open(flag.Args()[0])
		if err != nil {
			logrus.Fatal(err)
		}
		defer fh.Close()
		tarInput = fh
		logrus.Infof("using %q ...", fh.Name())
	}

	layers := dockerfile.LayerDatas{}
	rd := dockerfile.RepoData{}
	tr := tar.NewReader(tarInput)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err != io.EOF {
				logrus.Error(err)
			}
			break
		}
		if path.Base(hdr.Name) != "json" && path.Base(hdr.Name) != "repositories" {
			continue
		}
		buf, err := ioutil.ReadAll(tr)
		if err != nil {
			logrus.Fatal(err)
		}
		if path.Base(hdr.Name) == "json" {
			ld := dockerfile.LayerData{}
			if err := json.Unmarshal(buf, &ld); err != nil {
				logrus.Fatal(err)
			}
			logrus.Debugf("%#v", ld)
			layers = append(layers, &ld)
		} else if path.Base(hdr.Name) == "repositories" {
			if err := json.Unmarshal(buf, &rd); err != nil {
				logrus.Fatal(err)
			}
		}
	}
	// get IDs of images in this tar (from dockerfile.RepoData)
	rr := rd.References()
	logrus.Debugf("%#v", rr)

	// get these child dockerfile.LayerData with these IDs
	layers.BuildTrees()

	// since there could be more than one image in a "repositories" file
	dockerfiles := []dockerfile.Dockerfile{}
	for _, ref := range rr {
		for _, layer := range layers {
			if layer.ID == ref.ID {
				// layer here is a leaf image, with all parent nodes at layer.Parent
				df := dockerfile.Dockerfile{Ref: ref, Layers: dockerfile.LayerDatas{}}
				// walk up the parents
				curr := layer
				for {
					// add layer to the current Dockerfile
					df.Layers = append(df.Layers, curr)
					// quit if this layer is the parent
					if curr.ParentID == "" {
						break
					}
					curr = curr.Parent
				}
				// reverse the layers
				df.Layers.Reverse()
				dockerfiles = append(dockerfiles, df)
			}
		}
	}

	// build a reverse list of instructions from these child Layers
	// write these Dockerfile.XXX with these instructions
	tdir, err := ioutil.TempDir("", "docker-save-dockerfile.")
	if err != nil {
		logrus.Fatal(err)
	}

	for _, df := range dockerfiles {
		func() {
			fh, err := ioutil.TempFile(tdir, fmt.Sprintf("Dockerfile.%s.", df.Ref.ID))
			if err != nil {
				logrus.Errorf("%q: %s", df.Ref, err)
				return
			}
			defer fh.Close()

			if _, err := df.WriteTo(fh); err != nil {
				logrus.Errorf("%q: %s", df.Ref, err)
				return
			}
			logrus.Infof("Wrote %q to %q", df.Ref, fh.Name())
		}()
	}
}
