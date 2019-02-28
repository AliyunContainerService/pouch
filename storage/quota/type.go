package quota

// QMap defines the path set quota size and quota id.
type QMap struct {
	Source      string
	Destination string
	Expression  string
	Size        string
	QuotaID     uint32
}

// OverlayMount represents the parameters of overlay mount.
type OverlayMount struct {
	Merged string
	Lower  string
	Upper  string
	Work   string
}
