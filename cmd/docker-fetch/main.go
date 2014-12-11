package main

import (
	"fmt"
	"os"

	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/registry"
)

var (
	insecureRegistries = []string{"0.0.0.0/16"}
	timeout            = true
)

func init() {
	flag.BoolVar(&timeout, []string{"-timeout"}, timeout, "allow timeout on the registry session")
	opts.ListVar(&insecureRegistries, []string{"-insecure-registry"}, "Enable insecure communication with specified registries (no certificate verification for HTTPS and enable HTTP fallback) (e.g., localhost:5000 or 10.20.0.0/16) (default to 0.0.0.0/16)")
}

func main() {
	flag.Parse()
	var sessions map[string]*registry.Session
	for _, arg := range os.Args[1:] {
		host, imageName, err := registry.ResolveRepositoryName(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		e, err := registry.NewEndpoint(host, insecureRegistries)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("Pulling from %s\n", e)

		var session *registry.Session
		if s, ok := sessions[e.String()]; ok {
			session = s
		} else {
			// TODO(vbatts) obviously the auth and http factory shouldn't be nil here
			session, err = registry.NewSession(nil, nil, e, timeout)
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
		for _, img := range rd.ImgList {
			fmt.Printf("%#v\n", img)
		}
	}
}
