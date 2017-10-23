package randomid

import (
	"encoding/hex"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Generate generates a random string whose length is 64.
func Generate() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // This shouldn't happen
	}
	return hex.EncodeToString(b)
}
