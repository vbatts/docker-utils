package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/archive"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/vbatts/docker-utils/registry/fetch"
)

var (
	insecureRegistries = []string{"0.0.0.0/16"}
	timeout            = true
	debug              = len(os.Getenv("DEBUG")) > 0
	outputStream       = "-"
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	logrus.SetOutput(os.Stderr)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.WarnLevel)

	// XXX print a warning that this tool is not stable yet
	logrus.Warn("This tool is not stable yet, and should only be used for testing!")

	flag.BoolVar(&debug, []string{"D", "-debug"}, debug, "debugging output")
	flag.StringVar(&outputStream, []string{"o", "-output"}, outputStream, "output to file (default stdout)")
}

func main() {
	flag.Usage = func() {
		flag.PrintDefaults()
	}
	flag.Parse()
	if debug {
		os.Setenv("DEBUG", "1")
		logrus.SetLevel(logrus.DebugLevel)
	}
	if flag.NArg() == 0 {
		flag.Usage()
		logrus.Fatal("no image names provided")
	}

	// make temporary working directory
	tempFetchRoot, err := ioutil.TempDir("", "docker-fetch-")
	if err != nil {
		logrus.Fatal(err)
	}

	refs := []fetch.ImageRef{}
	for _, arg := range flag.Args() {
		ref := fetch.NewImageRef(arg)
		fmt.Fprintf(os.Stderr, "Pulling %s\n", ref)
		r := fetch.NewRegistry(ref.Host())

		layersFetched, err := r.FetchLayers(ref, tempFetchRoot)
		if err != nil {
			logrus.Errorf("failed pulling %s, skipping: %s", ref, err)
			continue
		}
		logrus.Debugf("fetched %d layers for %s", len(layersFetched), ref)
		refs = append(refs, ref)
	}

	// marshal the "repositories" file for writing out
	buf, err := fetch.FormatRepositories(refs...)
	if err != nil {
		logrus.Fatal(err)
	}
	fh, err := os.Create(filepath.Join(tempFetchRoot, "repositories"))
	if err != nil {
		logrus.Fatal(err)
	}
	if _, err = fh.Write(buf); err != nil {
		logrus.Fatal(err)
	}
	fh.Close()
	logrus.Debugf("%s", fh.Name())

	var output io.WriteCloser
	if outputStream == "-" {
		output = os.Stdout
	} else {
		output, err = os.Create(outputStream)
		if err != nil {
			logrus.Fatal(err)
		}
	}
	defer output.Close()

	if err = os.Chdir(tempFetchRoot); err != nil {
		logrus.Fatal(err)
	}
	tarStream, err := archive.Tar(".", archive.Uncompressed)
	if err != nil {
		logrus.Fatal(err)
	}
	if _, err = io.Copy(output, tarStream); err != nil {
		logrus.Fatal(err)
	}
}
