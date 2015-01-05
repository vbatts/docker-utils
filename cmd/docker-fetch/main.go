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

	log "github.com/Sirupsen/logrus"
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
	log.SetFormatter(&log.JSONFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	log.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)

	// XXX print a warning that this tool is not stable yet
	fmt.Fprintln(os.Stderr, "WARNING: this tool is not stable yet, and should only be used for testing!")

	flag.BoolVar(&timeout, []string{"t", "-timeout"}, timeout, "allow timeout on the registry session")
	flag.BoolVar(&debug, []string{"D", "-debug"}, debug, "debugging output")
	flag.StringVar(&outputStream, []string{"o", "-output"}, outputStream, "output to file (default stdout)")
	opts.ListVar(&insecureRegistries, []string{"-insecure-registry"}, "Enable insecure communication with specified registries (no certificate verification for HTTPS and enable HTTP fallback) (e.g., localhost:5000 or 10.20.0.0/16) (default to 0.0.0.0/16)")
}

// TODO rewrite this whole PoC
func main() {
	flag.Parse()
	if debug {
		log.SetLevel(log.DebugLevel)
	}

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

		indexEndpoint, err := registry.NewEndpoint(hostName, insecureRegistries)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Pulling %s:%s from %s\n", imageName, tagName, indexEndpoint)

		var session *registry.Session
		if s, ok := sessions[indexEndpoint.String()]; ok {
			session = s
		} else {
			// TODO(vbatts) obviously the auth and http factory shouldn't be nil here
			session, err = registry.NewSession(nil, nil, indexEndpoint, timeout)
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
		log.Debugf("rd: %#v", rd)

		// produce the "repositories" file for the archive
		if _, ok := repositories[imageName]; !ok {
			repositories[imageName] = graph.Repository{}
		}
		log.Debugf("repositories: %#v", repositories)

		if len(rd.Endpoints) == 0 {
			log.Fatalf("expected registry endpoints, but received none from the index")
		}

		tags, err := session.GetRemoteTags(rd.Endpoints, imageName, rd.Tokens)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if hash, ok := tags[tagName]; ok {
			repositories[imageName][tagName] = hash
		}
		log.Debugf("repositories: %#v", repositories)

		imgList, err := session.GetRemoteHistory(repositories[imageName][tagName], rd.Endpoints[0], rd.Tokens)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		log.Debugf("imgList: %#v", imgList)

		for _, imgID := range imgList {
			// pull layers and jsons
			buf, _, err := session.GetRemoteImageJSON(imgID, rd.Endpoints[0], rd.Tokens)
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
			log.Debugf("%s", fh.Name())

			tarRdr, err := session.GetRemoteImageLayer(imgID, rd.Endpoints[0], rd.Tokens, 0)
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
				log.Debugf("image layer for %q is not gzipped", imgID)
				// the archive may not be gzipped, so just copy the stream
				if _, err = io.Copy(fh, tarRdr); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			} else {
				// no error, so gzip decompress the stream
				if _, err = io.Copy(fh, gzRdr); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				if err = gzRdr.Close(); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
			if err = tarRdr.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			if err = fh.Close(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			log.Debugf("%s", fh.Name())
		}
	}

	// marshal the "repositories" file for writing out
	log.Debugf("repositories: %q", repositories)
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
	log.Debugf("%s", fh.Name())

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
