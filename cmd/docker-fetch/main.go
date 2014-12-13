package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/graph"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/archive"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/registry"
)

var (
	insecureRegistries = []string{"0.0.0.0/16"}
	timeout            = true
	debug              = len(os.Getenv("DEBUG")) > 0
	outputStream       = "-"
)

func init() {
	// XXX print a warning that this tool is not stable yet
	fmt.Fprintln("WARNING: this tool is not stable yet, and should only be used for testing!")

	flag.BoolVar(&timeout, []string{"t", "-timeout"}, timeout, "allow timeout on the registry session")
	flag.BoolVar(&debug, []string{"D", "-debug"}, debug, "debugging output")
	flag.StringVar(&outputStream, []string{"o", "-output"}, outputStream, "output to file (default stdout)")
	opts.ListVar(&insecureRegistries, []string{"-insecure-registry"}, "Enable insecure communication with specified registries (no certificate verification for HTTPS and enable HTTP fallback) (e.g., localhost:5000 or 10.20.0.0/16) (default to 0.0.0.0/16)")
}

func main() {
	flag.Parse()
	var (
		sessions     map[string]*registry.Session
		repositories = map[string]graph.Repository{}
	)

	// make tempDir
	tempDir, err := ioutil.TempDir("", "docker-fetch-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	for _, arg := range flag.Args() {
		var (
			hostName, imageName, tagName string
			err                          error
		)

		hostName, imageName, err = registry.ResolveRepositoryName(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		// set up image and tag
		if strings.Contains(imageName, ":") {
			chunks := strings.SplitN(imageName, ":", 2)
			imageName = chunks[0]
			tagName = chunks[1]
		} else {
			tagName = "latest"
		}

		endpoint, err := registry.NewEndpoint(hostName, insecureRegistries)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Pulling %s:%s from %s\n", imageName, tagName, endpoint)

		var session *registry.Session
		if s, ok := sessions[endpoint.String()]; ok {
			session = s
		} else {
			// TODO(vbatts) obviously the auth and http factory shouldn't be nil here
			session, err = registry.NewSession(nil, nil, endpoint, timeout)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		rd, err := session.GetRepositoryData(imageName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if debug {
			fmt.Fprintf(os.Stderr, "%#v\n", rd)
		}
		/*
			for _, img := range rd.ImgList {
				fmt.Fprintf(os.Stderr, "%#v\n", img)
			}
		*/

		// produce the "repositories" file for the archive
		if _, ok := repositories[imageName]; !ok {
			repositories[imageName] = graph.Repository{}
		}

		tags, err := session.GetRemoteTags([]string{endpoint.String()}, imageName, rd.Tokens)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if hash, ok := tags[tagName]; ok {
			repositories[imageName][tagName] = hash
		}

		imgList, err := session.GetRemoteHistory(repositories[imageName][tagName], endpoint.String(), rd.Tokens)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		for _, imgID := range imgList {
			// pull layers and jsons
			buf, _, err := session.GetRemoteImageJSON(imgID, endpoint.String(), rd.Tokens)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err = os.MkdirAll(filepath.Join(tempDir, imgID), 0755); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fh, err := os.Create(filepath.Join(tempDir, imgID, "json"))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if _, err = fh.Write(buf); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fh.Close()
			if debug {
				fmt.Fprintln(os.Stderr, fh.Name())
			}

			tarRdr, err := session.GetRemoteImageLayer(imgID, endpoint.String(), rd.Tokens, 0)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			fh, err = os.Create(filepath.Join(tempDir, imgID, "layer.tar"))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			// the body is usually compressed
			gzRdr, err := gzip.NewReader(tarRdr)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if _, err = io.Copy(fh, gzRdr); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err = gzRdr.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err = tarRdr.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err = fh.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if debug {
				fmt.Fprintln(os.Stderr, fh.Name())
			}
		}
	}

	// marshal the "repositories" file for writing out
	if debug {
		fmt.Fprintf(os.Stderr, "%q", repositories)
	}
	buf, err := json.Marshal(repositories)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fh, err := os.Create(filepath.Join(tempDir, "repositories"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err = fh.Write(buf); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fh.Close()
	if debug {
		fmt.Fprintln(os.Stderr, fh.Name())
	}

	var output io.WriteCloser
	if outputStream == "-" {
		output = os.Stdout
	} else {
		output, err = os.Create(outputStream)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	defer output.Close()

	if err = os.Chdir(tempDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	tarStream, err := archive.Tar(".", archive.Uncompressed)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if _, err = io.Copy(output, tarStream); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
