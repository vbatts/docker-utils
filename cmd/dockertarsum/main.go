package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vbatts/docker-utils/opts"
	"github.com/vbatts/docker-utils/sum"
	"github.com/vbatts/docker-utils/version"
)

func main() {
	var (
		failedChecks = []bool{}
	)
	flag.Parse()

	if *flVersion {
		fmt.Fprintf(os.Stderr, "%s - %s\n", os.Args[0], version.VERSION)
		os.Exit(0)
	}
	tsVersion, err := sum.DetermineVersion(*flTarsumVersion)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	checks, err := sum.LoadCheckFiles(flChecks.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	isIncluded := false
	for _, v := range checks.Versions() {
		if v == tsVersion {
			isIncluded = true
			break
		}
	}
	if len(checks) != 0 && !isIncluded {
		fmt.Fprintf(os.Stderr, "ERROR: the TarSum version %q is not included in the check file\n", tsVersion)
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		if *flStream {
			var hashes map[string]string
			var err error
			if !*flRootTar {
				// this filehandle `fh` is a tar of tars from `docker save`
				if hashes, err = sum.SumAllDockerSaveVersioned(os.Stdin, tsVersion); err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
					os.Exit(1)
				}
			} else {
				// this filehandle `fh` is a rootfs
				hash, err := sum.SumTarLayerVersioned(os.Stdin, nil, nil, tsVersion)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
					os.Exit(1)
				}
				hashes = map[string]string{"-": hash}
			}
			for id, hash := range hashes {
				if len(checks) == 0 {
					// nothing to check against, just print the hash
					fmt.Printf("%s%s-:%s\n", hash, sum.DefaultSpacer, id)
				} else {
					// check the sum against the checks available
					check := checks.Get(id)
					if check == nil {
						fmt.Fprintf(os.Stderr, "WARNING: no check found for ID [%s]\n", id)
						continue
					}
					check.Seen = true // so can print NOT FOUND IDs
					var result string
					if check.Hash != hash {
						result = "FAILED"
						failedChecks = append(failedChecks, false)
					} else {
						result = "OK"
					}
					fmt.Printf("%s:%s%s\n", id, sum.DefaultSpacer, result)
				}
			}
		} else {
			// maybe the actual layer.tar ? and json? or image name and we'll call a docker daemon?
			fmt.Fprintln(os.Stderr, "ERROR: not implemented yet")
			os.Exit(2)
		}
	}

	for _, arg := range flag.Args() {
		fh, err := os.Open(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
			os.Exit(1)
		}
		if *flStream {
			var hashes map[string]string
			if !*flRootTar {
				// this filehandle `fh` is a tar of tars from `docker save`
				if hashes, err = sum.SumAllDockerSaveVersioned(fh, tsVersion); err != nil {
					fmt.Printf("ERROR: %s\n", err)
					os.Exit(1)
				}
			} else {
				// this filehandle `fh` is a rootfs
				hash, err := sum.SumTarLayerVersioned(fh, nil, nil, tsVersion)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
					os.Exit(1)
				}
				hashes = map[string]string{arg: hash}
			}
			for id, hash := range hashes {
				if len(checks) == 0 {
					fmt.Printf("%s%s%s:%s\n", hash, sum.DefaultSpacer, arg, id)
				} else {
					// check the sum against the checks available
					check := checks.Get(id)
					if check == nil {
						fmt.Fprintf(os.Stderr, "WARNING: no check found for ID [%s]\n", id)
						continue
					}
					check.Seen = true // so can print NOT FOUND IDs
					var result string
					if check.Hash != hash {
						result = "FAILED"
						failedChecks = append(failedChecks, false)
					} else {
						result = "OK"
					}
					fmt.Printf("%s:%s%s\n", id, sum.DefaultSpacer, result)
				}
			}
		} else {
			// maybe the actual layer.tar ? and json? or image name and we'll call a docker daemon?
			fmt.Fprintln(os.Stderr, "ERROR: not implemented yet")
			os.Exit(2)
		}
	}

	// print out the rest of the checks info
	if len(checks) > 0 {
		for _, c := range checks {
			if !c.Seen {
				fmt.Printf("%s:%sNOT FOUND\n", c.Id, sum.DefaultSpacer)
			}
		}
		if len(failedChecks) > 0 {
			fmt.Printf("%s: WARNING: %d computed checksums did NOT match\n", os.Args[0], len(failedChecks))
			os.Exit(1)
		}
	}
}

var (
	flChecks        = opts.List{}
	flTarsumVersion = flag.String("t", "Version1", "Which version of the tarsum checksum to use")
	flStream        = flag.Bool("s", true, "read FILEs (or stdin) as the output of `docker save` (this is default)")
	flVersion       = flag.Bool("v", false, "show version")
	flRootTar       = flag.Bool("r", false, "treat the tar(s) root filesystem archives (not a tar of layers)")
)

func init() {
	flag.Var(&flChecks, "c", "read TarSums from the FILEs (or stdin) and check them")
}
