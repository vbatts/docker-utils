package fetch

import "strings"

// NewImageRef constructs a reference to a distributable container image,
// like my.registry.com/vbatts/myapp:stable
func NewImageRef(name string) ImageRef {
	return &imageRef{orig: name}
}

// ImageRef provides access to attributes and data regarding a distributable
// container image
type ImageRef interface {
	Hoster                // the hostname from the image reference
	Name() string         // the name (according to docker's formatting) of the image reference
	ID() string           // image's ID, if available
	SetID(string)         // set the ID for the image reference
	Ancestry() []string   // List of ancestor IDs, if available
	SetAncestry([]string) // set the ancestry for the image reference
	Tag() string          // the tag (according to docker's formatting) of the image reference
	Digest() string       // image's digest, if available
	String() string       // pretty print the image's reference
}

type imageRef struct {
	orig     string
	name     string
	tag      string
	digest   string
	id       string
	ancestry []string
}

func (ir imageRef) Host() string {
	// if there are 2 or more slashes and the first element includes a period
	if strings.Count(ir.orig, "/") > 0 {
		// first element
		el := strings.Split(ir.orig, "/")[0]
		// it looks like an address or is localhost
		if strings.Contains(el, ".") || el == "localhost" || strings.Contains(el, ":") {
			return el
		}
	}
	return DefaultHubNamespace
}

func (ir imageRef) ID() string {
	return ir.id
}
func (ir *imageRef) SetID(id string) {
	ir.id = id
}

func (ir imageRef) Ancestry() []string {
	return ir.ancestry
}
func (ir *imageRef) SetAncestry(ids []string) {
	ir.ancestry = make([]string, len(ids))
	for i := range ids {
		ir.ancestry[i] = ids[i]
	}
}
func (ir imageRef) Name() string {
	// trim off the hostname plus the slash
	name := strings.TrimPrefix(ir.orig, ir.Host()+"/")

	// check for any tags
	count := strings.Count(name, ":")
	if count == 0 {
		return name
	}
	if count == 1 {
		return strings.Split(name, ":")[0]
	}
	return ""
}
func (ir imageRef) Tag() string {
	if ir.tag != "" {
		return ir.tag
	}
	count := strings.Count(ir.orig, ":")
	if count == 0 {
		return DefaultTag
	}
	if c := strings.Count(ir.orig, "/"); c > 0 {
		el := strings.Split(ir.orig, "/")[c]
		if strings.Contains(el, ":") {
			return strings.Split(el, ":")[1]
		}
		return DefaultTag
	}
	if count == 1 {
		return strings.Split(ir.orig, ":")[1]
	}
	return ""
}

func (ir imageRef) Digest() string {
	if ir.digest != "" {
		return ir.digest
	}
	return ""
}

func (ir imageRef) String() string {
	return ir.Host() + "/" + ir.Name() + ":" + ir.Tag()
}
