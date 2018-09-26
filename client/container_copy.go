package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/alibaba/pouch/apis/types"
)

// ContainerStatPath returns Stat information about a path inside the container filesystem.
func (client *APIClient) ContainerStatPath(ctx context.Context, name string, path string) (types.ContainerPathStat, error) {
	headers := make(map[string][]string)
	headers["SRCPATH"] = []string{path}

	urlStr := fmt.Sprintf("/containers/%s/archive", name)

	response, err := client.head(ctx, urlStr, nil, headers)
	if err != nil {
		return types.ContainerPathStat{}, err
	}

	if err != nil {
		return types.ContainerPathStat{}, err
	}
	ensureCloseReader(response)
	return getContainerPathStatFromHeader(response.Header)
}

// CopyFromContainer gets the content from the container and returns it as a Reader
// to manipulate it in the host. It's up to the caller to close the reader.
func (client *APIClient) CopyFromContainer(ctx context.Context, container, srcPath string) (io.ReadCloser, types.ContainerPathStat, error) {
	headers := make(map[string][]string)
	headers["SRCPATH"] = []string{srcPath}

	apiPath := fmt.Sprintf("/containers/%s/archive", container)
	response, err := client.get(ctx, apiPath, nil, headers)
	if err != nil {
		return nil, types.ContainerPathStat{}, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, types.ContainerPathStat{}, fmt.Errorf("unexpected status code from daemon: %d", response.StatusCode)
	}

	// In order to get the copy behavior right, we need to know information
	// about both the source and the destination. The response headers include
	// stat info about the source that we can use in deciding exactly how to
	// copy it locally. Along with the stat info about the local destination,
	// we have everything we need to handle the multiple possibilities there
	// can be when copying a file/dir from one location to another file/dir.
	stat, err := getContainerPathStatFromHeader(response.Header)
	if err != nil {
		return nil, stat, fmt.Errorf("unable to get resource stat from response: %s", err)
	}
	return response.Body, stat, err
}

// CopyToContainer copies content into the container filesystem.
func (client *APIClient) CopyToContainer(ctx context.Context, container, path string, content io.Reader, options types.CopyToContainerOptions) error {
	query := url.Values{}
	if !options.AllowOverwriteDirWithFile {
		query.Set("noOverwriteDirNonDir", "true")
	}
	query.Set("SRCPATH", filepath.ToSlash(path))

	apiPath := fmt.Sprintf("/containers/%s/archive", container)

	response, err := client.putRaw(ctx, apiPath, query, content, nil)
	if err != nil {
		return err
	}
	ensureCloseReader(response)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from daemon: %d", response.StatusCode)
	}

	return nil
}

func getContainerPathStatFromHeader(header http.Header) (types.ContainerPathStat, error) {
	var stat types.ContainerPathStat

	encodedStat := header.Get("X-Pouch-Container-Path-Stat")
	statDecoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(encodedStat))

	err := json.NewDecoder(statDecoder).Decode(&stat)
	if err != nil {
		err = fmt.Errorf("unable to decode container path stat header: %s", err)
	}

	return stat, err
}
