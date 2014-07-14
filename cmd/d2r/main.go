package main

import (
	"flag"
	"fmt"
	"github.com/vbatts/docker-utils/registry"
	//"github.com/vbatts/docker-utils/version"
	"os"
)

var (
	flOutdir = flag.String("o", "./static/", "directory to land the output registry files")
	//flVersion = flag.Bool("v", false, "show version")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: %s [OPTIONS] <file.tar|->\n  (where '-' is from stdin)\n", os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	//if *flVersion {
	//fmt.Printf("%s - %s\n", os.Args[0], version.VERSION)
	//os.Exit(0)
	//}

	reg := registry.Registry{Path: *flOutdir}
	err := reg.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, arg := range flag.Args() {
		if arg == "-" {
			if err = registry.ExtractTar(&reg, os.Stdin); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		} else {
			fh, err := os.Open(arg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer fh.Close()
			if err := registry.ExtractTar(&reg, fh); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
	}
}
