package credential

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	assert := assert.New(t)
	type authTest struct {
		username string
		password string
		pass     bool
	}
	for _, auth := range []authTest{
		{
			username: "username1",
			password: "password1",
			pass:     true,
		},
		{
			pass: false,
		},
		{
			username: "",
			pass:     false,
		},
		{
			password: "",
			pass:     false,
		},
		{
			username: "231dsda",
			password: "asddwqwe333!#",
			pass:     true,
		},
	} {
		authStr := encodeAuth(auth.username, auth.password)
		if auth.pass {
			t.Logf("username %s and password %s encode: %s", auth.username, auth.password, authStr)
			u, p, err := decodeAuth(authStr)
			assert.NoError(err)
			assert.Equal(u, auth.username)
			assert.Equal(p, auth.password)
		} else {
			assert.Equal("", authStr)
		}

	}
}
