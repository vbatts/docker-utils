package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vbatts/docker-utils/opts"
	"github.com/vbatts/docker-utils/sum"
	"github.com/vbatts/docker-utils/version"
)

func main() {
	var (
		checks       Checks
		failedChecks = []bool{}
	)
	flag.Parse()

	if flVersion {
		fmt.Printf("%s - %s\n", os.Args[0], version.VERSION)
		os.Exit(0)
	}

	if len(flChecks.Args) > 0 {
		checks = Checks{}
		for _, c := range flChecks.Args {
			fh, err := os.Open(c)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(1)
			}
			newChecks, err := ReadChecks(fh)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(1)
			}
			checks = append(checks, newChecks...)
		}
	}

	if flag.NArg() == 0 {
		if flStream {
			// assumption is this is stdin from `docker save`
			hashes, err := sum.SumAllDockerSave(os.Stdin)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(1)
			}
			for id, hash := range hashes {
				if len(checks) == 0 {
					// nothing to check against, just print the hash
					fmt.Printf("%s%s-:%s\n", hash, DefaultSpacer, id)
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
					fmt.Printf("%s:%s%s\n", id, DefaultSpacer, result)
				}
			}
		} else {
			// maybe the actual layer.tar ? and json? or image name and we'll call a docker daemon?
			fmt.Println("ERROR: not implemented yet")
			os.Exit(2)
		}
	}

	for _, arg := range flag.Args() {
		fh, err := os.Open(arg)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
		if flStream {
			// assumption is this is a tar from `docker save`
			hashes, err := sum.SumAllDockerSave(fh)
			if err != nil {
				fmt.Printf("ERROR: %s\n", err)
				os.Exit(1)
			}
			for id, hash := range hashes {
				if len(checks) == 0 {
					fmt.Printf("%s%s%s:%s\n", hash, DefaultSpacer, arg, id)
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
					fmt.Printf("%s:%s%s\n", id, DefaultSpacer, result)
				}
			}
		} else {
			// maybe the actual layer.tar ? and json? or image name and we'll call a docker daemon?
			fmt.Println("ERROR: not implemented yet")
			os.Exit(2)
		}
	}
	for _, c := range checks {
		if !c.Seen {
			fmt.Printf("%s:%sNOT FOUND\n", c.Id, DefaultSpacer)
		}
	}
	if len(failedChecks) > 0 {
		fmt.Printf("%s: WARNING: %d computed checksums did NOT match\n", os.Args[0], len(failedChecks))
		os.Exit(1)
	}
}

// ReadChecks takes the input and loads the hash/id to be checked
func ReadChecks(input io.Reader) (Checks, error) {
	rdr := bufio.NewReader(input)
	checks := Checks{}
	for {
		line, err := rdr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return checks, err
		}
		// skip non-tarsums
		if !strings.HasPrefix(line, "tarsum+") {
			continue
		}
		// XXX parse the line
		// tarsum+sha256:7b0ade22d5bba35d1e88389c005376f441e7d83bf5f363f2d7c70be9286163aa  ./busybox.tar:120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16
		chunks := strings.SplitN(line, DefaultSpacer, 2)
		sum, source := chunks[0], chunks[1]
		i := strings.LastIndex(source, ":")
		checks = append(checks, Check{Hash: sum, Source: source[:i], Id: strings.TrimSpace(source[i+1:])})
	}
	return checks, nil
}

type Check struct {
	Id     string
	Source string
	Hash   string
	Seen   bool
}

type Checks []Check

func (c Checks) Get(id string) *Check {
	for i := range c {
		if id == c[i].Id {
			return &c[i]
		}
	}
	return nil
}

var (
	flChecks  = opts.List{}
	flStream  = false
	flVersion = false
)

func init() {
	flag.Var(&flChecks, "c", "read TarSums from the FILEs (or stdin) and check them")
	flag.BoolVar(&flStream, "s", true, "read FILEs (or stdin) as the output of `docker save` (this is default)")
	flag.BoolVar(&flVersion, "v", false, "show version")
}

const DefaultSpacer = "  "
