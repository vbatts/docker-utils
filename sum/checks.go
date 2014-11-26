package sum

import (
	"bufio"
	"io"
	"strings"

	"github.com/docker/docker/pkg/tarsum"
)

const DefaultSpacer = "  "

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
		v, err := tarsum.GetVersionFromTarsum(line)
		if err != nil {
			continue
		}
		// tarsum+sha256:7b0ade22d5bba35d1e88389c005376f441e7d83bf5f363f2d7c70be9286163aa  ./busybox.tar:120e218dd395ec314e7b6249f39d2853911b3d6def6ea164ae05722649f34b16
		chunks := strings.SplitN(line, DefaultSpacer, 2)
		sum, source := chunks[0], chunks[1]
		i := strings.LastIndex(source, ":")
		checks = append(checks, Check{Hash: sum, Source: source[:i], Id: strings.TrimSpace(source[i+1:]), Version: v})
	}
	return checks, nil
}

type Check struct {
	Id      string
	Source  string
	Hash    string
	Seen    bool
	Version tarsum.Version
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
