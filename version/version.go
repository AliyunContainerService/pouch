package version

import "time"

// Version represents the version of pouchd.
const Version = "0.1.0-dev"

// BuildTime is the time when this binary of daemon is built
// FIXME: this is dynamical. We need a fixed build time.
var BuildTime = time.Now().Format(time.RFC3339Nano)

// APIVersion means the api version daemon serves
var APIVersion = "vx.y.z"
