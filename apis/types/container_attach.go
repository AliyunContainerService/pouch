package types

import (
	"net/http"
)

// AttachConfig wraps some infos of attaching.
type AttachConfig struct {
	Hijack  http.Hijacker
	Stdin   bool
	Stdout  bool
	Stderr  bool
	Upgrade bool
}
