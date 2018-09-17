package stream

import (
	"net/http"
)

// Router exports a set of CRI Stream Server's handlers.
// We could reuse the pouchd's http server to handle
// the Stream Server's requests, so pouchd only has to
// export one port.
type Router interface {
	ServeExec(w http.ResponseWriter, r *http.Request)
	ServeAttach(w http.ResponseWriter, r *http.Request)
	ServePortForward(w http.ResponseWriter, r *http.Request)
}
