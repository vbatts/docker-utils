package opts

import (
	"fmt"
	"strings"
)

type List struct {
	Args []string
}

func (lo *List) Set(arg string) error {
	lo.Args = append(lo.Args[:], arg)
	return nil
}

func (lo List) Get() []string {
	return lo.Args
}

func (lo List) String() string {
	return fmt.Sprintf("[%s]", strings.Join(lo.Args, ", "))
}
