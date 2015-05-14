package dockerfile

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
)

type Dockerfile struct {
	// Ordered list of Layers
	Layers LayerDatas
	// For further information
	Ref RepoReference
}

// WriteTo will render the instructions from this Dockerfile to the provided
// io.Writer. This satisfies the io.WriterTo interface.
func (d Dockerfile) WriteTo(w io.Writer) (n int64, err error) {
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	i, err := bw.WriteString(fmt.Sprintf("## RECREATED FROM IMAGE ON %s\n", time.Now().Format(time.RFC3339)))
	if err != nil {
		return n, err
	}
	n += int64(i)
	i, err = bw.WriteString(fmt.Sprintf("## %s:%s (%s)\n\n", d.Ref.Name, d.Ref.Tag, d.Ref.ID))
	if err != nil {
		return n, err
	}
	n += int64(i)
	for _, layer := range d.Layers {
		i, err = bw.WriteString(layer.DockerfileInstruction().String() + "\n")
		if err != nil {
			return n, err
		}
		n += int64(i)
	}
	return n, nil
}

type DockerfileInstruction struct {
	Comment string
	Command string
	Args    string
}

// String will render a comment safe string for output into a Dockerfile
func (di DockerfileInstruction) String() string {
	return fmt.Sprintf("# %s\n%s %s\n", di.Comment, strings.ToUpper(di.Command), di.Args)
}

// RepoData consists of map[repoName]TagMap, for the easy json marshal
type RepoData map[string]TagMap

func (rd RepoData) References() []RepoReference {
	rr := []RepoReference{}
	for repoName, tm := range rd {
		for tagName, imageID := range tm {
			rr = append(rr, RepoReference{Name: repoName, Tag: tagName, ID: imageID})
		}
	}
	return rr
}

// TagMap consists of map[tagname]imageID, for the easy json marshal
type TagMap map[string]string

type RepoReference struct {
	Name string
	Tag  string
	ID   string
}

type LayerDatas []*LayerData

func (ld *LayerDatas) Reverse() {
	for i, j := 0, len(*ld)-1; i < j; i, j = i+1, j-1 {
		(*ld)[i], (*ld)[j] = (*ld)[j], (*ld)[i]
	}
}

// walk the list of nodes, and build the parent child relationships
func (ld *LayerDatas) BuildTrees() {
	for _, image := range *ld {
		if (*image).ParentID == "" {
			continue
		}
		for _, parent := range *ld {
			if (*image).ParentID == (*parent).ID {
				(*image).Parent = parent
			}
		}
	}
}

type LayerData struct {
	ID              string          `json:"id"`
	ParentID        string          `json:"parent"`
	Comment         string          `json:"comment,omitempty"`
	Created         time.Time       `json:"created"`
	Author          string          `json:"author,omitempty"`
	ContainerConfig ContainerConfig `json:"container_config,omitempty"`

	Parent *LayerData
}

func (ld LayerData) DockerfileInstruction() DockerfileInstruction {
	di := DockerfileInstruction{}

	comments := []string{
		fmt.Sprintf("Created: %s", ld.Created),
		fmt.Sprintf("ID: %s", ld.ID),
	}
	if ld.Author != "" {
		comments = append(comments, fmt.Sprintf("Author: %q", ld.Author))
	}
	if ld.Comment != "" {
		comments = append(comments, fmt.Sprintf("Comment: %q", ld.Comment))
	}

	if ld.ParentID == "" {
		// this is as good as it gets for FROM :-\
		di.Command = "FROM"
		di.Args = ld.ID
	} else {
		if len(ld.ContainerConfig.Cmd) == 3 && strings.HasPrefix(ld.ContainerConfig.Cmd[2], NopPrefix) {
			// actual command
			trim := strings.TrimSpace(strings.TrimPrefix(ld.ContainerConfig.Cmd[2], NopPrefix))
			s := strings.SplitN(trim, " ", 2)
			if len(s) != 2 {
				logrus.Warnf("expected two items, but got %s", s)
			} else {
				di.Command = s[0]
				di.Args = s[1]
			}
		} else {
			// RUN
			di.Command = "RUN"
			di.Args = strings.Join(ld.ContainerConfig.Cmd, " ")
		}
	}
	// put this last, as others may have been added
	di.Comment = strings.Join(comments, "; ")
	return di
}

var NopPrefix = "#(nop)"

type ContainerConfig struct {
	Cmd []string `json:"Cmd"`
}
