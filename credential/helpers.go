package credential

import (
	"encoding/base64"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func encodeAuth(username, password string) string {
	if username == "" || password == "" {
		return ""
	}

	str := username + ":" + password
	msgBytes := []byte(str)
	encodedBytes := make([]byte, base64.StdEncoding.EncodedLen(len(msgBytes)))
	base64.StdEncoding.Encode(encodedBytes, msgBytes)
	return string(encodedBytes)
}

func decodeAuth(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", nil
	}

	decodeLen := base64.StdEncoding.DecodedLen(len(authStr))
	decodeBytes := make([]byte, decodeLen)
	authBytes := []byte(authStr)
	n, err := base64.StdEncoding.Decode(decodeBytes, authBytes)
	if err != nil {
		return "", "", err
	}
	if n > decodeLen {
		return "", "", fmt.Errorf("failed to decode auth %s", authStr)
	}

	splits := strings.Split(string(decodeBytes), ":")
	if len(splits) != 2 {
		return "", "", fmt.Errorf("invalid credential config")
	}

	password := strings.Trim(splits[1], "\x00")
	return splits[0], password, nil
}

func homedir() string {
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	}

	return os.Getenv(env)
}

func convertHost(addr string) string {
	if strings.HasPrefix(addr, "http://") {
		addr = strings.TrimPrefix(addr, "http://")
	}
	if strings.HasPrefix(addr, "https://") {
		addr = strings.TrimPrefix(addr, "https://")
	}

	splits := strings.SplitN(addr, "/", 2)
	return splits[0]
}
