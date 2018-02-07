package reference

// Reference represents image name which may include domain/name:tag or only digest.
type Reference interface {
	String() string
}

// Named is an object which has full name.
type Named interface {
	Reference
	Name() string
}

// Tagged is an Named object contains tag.
type Tagged interface {
	Named
	Tag() string
}

// Digested is an object which is digest.
type Digested interface {
	Named
	Digest() string
}

// namedReference represents the image short ID or Name.
type namedReference struct {
	name string
}

func (n namedReference) Name() string {
	return n.name
}

func (n namedReference) String() string {
	return n.name
}

// taggedReference represents the image Name:tag.
type taggedReference struct {
	Named
	tag string
}

func (t taggedReference) Tag() string {
	return t.tag
}

func (t taggedReference) String() string {
	return t.Name() + ":" + t.tag
}

// taggedReference represents the image digest information.
type digestReference struct {
	Named
	digest string
}

func (d digestReference) String() string {
	return d.Name() + "@" + d.digest
}

func (d digestReference) Digest() string {
	return d.digest
}
