package client

import (
	"net/url"
	"path"
)

// JoinURL is used to check "api" and join url.
func JoinURL(api string, s ...string) (string, error) {
	url, err := url.Parse(api)
	if err != nil {
		return "", err
	}
	p := []string{url.Path}
	p = append(p, s...)
	url.Path = path.Join(p...)

	return url.String(), nil
}
