package fetch

import "strings"

func NewImageRef(name string) *ImageRef {
	return &ImageRef{orig: name}
}

type ImageRef struct {
	orig     string
	name     string
	tag      string
	digest   string
	id       string
	ancestry []string
}

func (ir ImageRef) Host() string {
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

func (ir ImageRef) ID() string {
	return ir.id
}
func (ir *ImageRef) SetID(id string) {
	ir.id = id
}

func (ir ImageRef) Ancestry() []string {
	return ir.ancestry
}
func (ir *ImageRef) SetAncestry(ids []string) {
	ir.ancestry = make([]string, len(ids))
	for i := range ids {
		ir.ancestry[i] = ids[i]
	}
}
func (ir ImageRef) Name() string {
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
func (ir ImageRef) Tag() string {
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
		} else {
			return DefaultTag
		}
	}
	if count == 1 {
		return strings.Split(ir.orig, ":")[1]
	}
	return ""
}

func (ir ImageRef) Digest() string {
	if ir.digest != "" {
		return ir.digest
	}
	return ""
}

func (ir ImageRef) String() string {
	return ir.Host() + "/" + ir.Name() + ":" + ir.Tag()
}
