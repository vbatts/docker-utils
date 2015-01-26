package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/graph"
	"github.com/docker/docker/pkg/archive"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/parsers"
	"github.com/docker/docker/registry"
)

var (
	insecureRegistries = []string{"0.0.0.0/16"}
	timeout            = true
	debug              = len(os.Getenv("DEBUG")) > 0
	outputStream       = "-"
	rOptions           = &registry.Options{}
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

	rOptions.InstallFlags()
}

// TODO rewrite this whole PoC
func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
	if debug {
		os.Setenv("DEBUG", "1")
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

	sc := registry.NewServiceConfig(rOptions)

	for _, arg := range flag.Args() {
		remote, tagName := parsers.ParseRepositoryTag(arg)
		if tagName == "" {
			tagName = "latest"
		}

		repInfo, err := sc.NewRepositoryInfo(remote)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		log.Debugf("%#v %q\n", repInfo, tagName)

		idx, err := repInfo.GetEndpoint()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Pulling %s:%s from %s\n", repInfo.RemoteName, tagName, idx)

		var session *registry.Session
		if s, ok := sessions[idx.String()]; ok {
			session = s
		} else {
			// TODO(vbatts) obviously the auth and http factory shouldn't be nil here
			session, err = registry.NewSession(nil, nil, idx, timeout)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}

		rd, err := session.GetRepositoryData(repInfo.RemoteName)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		log.Debugf("rd: %#v", rd)

		// produce the "repositories" file for the archive
		if _, ok := repositories[repInfo.RemoteName]; !ok {
			repositories[repInfo.RemoteName] = graph.Repository{}
		}
		log.Debugf("repositories: %#v", repositories)

		if len(rd.Endpoints) == 0 {
			log.Fatalf("expected registry endpoints, but received none from the index")
		}

		tags, err := session.GetRemoteTags(rd.Endpoints, repInfo.RemoteName, rd.Tokens)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if hash, ok := tags[tagName]; ok {
			repositories[repInfo.RemoteName][tagName] = hash
		}
		log.Debugf("repositories: %#v", repositories)

		imgList, err := session.GetRemoteHistory(repositories[repInfo.RemoteName][tagName], rd.Endpoints[0], rd.Tokens)
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
