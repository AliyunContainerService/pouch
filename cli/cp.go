package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/pkg/archive"
	"github.com/spf13/cobra"
)

// createDescription is used to describe create command in detail and auto generate command doc.
var copyDescription = "Copy files/folders between a container and the local filesystem\n" +
	"\nUse '-' as the source to read a tar archive from stdin\n" +
	"and extract it to a directory destination in a container.\n" +
	"Use '-' as the destination to stream a tar archive of a\n" +
	"container source to stdout."

// CopyCommand use to implement 'copy' command, it copy files between host and container.
type CopyCommand struct {
	*container
	baseCommand
}

type copyOptions struct {
	source      string
	destination string
}

// Init initialize copy command.
func (cc *CopyCommand) Init(c *Cli) {
	var opts copyOptions

	cc.cli = c
	cc.cmd = &cobra.Command{
		Use:   "cp [OPTIONS] CONTAINER:SRC_PATH DEST_PATH|-\n  pouch cp [OPTIONS] SRC_PATH|- CONTAINER:DEST_PATH",
		Short: "Copy files/folders between a container and the local filesystem",
		Long:  copyDescription,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "" {
				return fmt.Errorf("source can not be empty")
			}
			if args[1] == "" {
				return fmt.Errorf("destination can not be empty")
			}

			opts.source = args[0]
			opts.destination = args[1]
			return cc.runCopy(opts)
		},
		Example: copyExample(),
	}

	cc.addFlags()
}

// addFlags adds flags for specific command.
func (cc *CopyCommand) addFlags() {
	flagSet := cc.cmd.Flags()
	flagSet.SetInterspersed(false)

	// TODO: add more flags here
}

func splitCpArg(arg string) (container, path string) {
	if filepath.IsAbs(arg) {
		// Explicit local absolute path, e.g., `/foo`.
		return "", arg
	}

	parts := strings.SplitN(arg, ":", 2)

	if len(parts) == 1 {
		return "", arg
	}

	return parts[0], parts[1]
}

type copyDirection int

const (
	fromContainer copyDirection = 1 << iota
	toContainer
	acrossContainers = fromContainer | toContainer
)

// runCopy is the entry of Copy command.
func (cc *CopyCommand) runCopy(opts copyOptions) error {
	srcContainer, srcPath := splitCpArg(opts.source)
	dstContainer, dstPath := splitCpArg(opts.destination)

	var direction copyDirection
	if srcContainer != "" {
		direction |= fromContainer
	}
	if dstContainer != "" {
		direction |= toContainer
	}

	ctx := context.Background()

	switch direction {
	case fromContainer:
		return copyFromContainer(ctx, cc.cli, srcContainer, srcPath, dstPath)
	case toContainer:
		return copyToContainer(ctx, cc.cli, srcPath, dstContainer, dstPath)
	case acrossContainers:
		// Copying between containers isn't supported.
		return fmt.Errorf("copying between containers is not supported")
	default:
		// User didn't specify any container.
		return fmt.Errorf("must specify at least one container source")
	}
}

func resolveLocalPath(localPath string) (absPath string, err error) {
	if absPath, err = filepath.Abs(localPath); err != nil {
		return
	}

	return archive.PreserveTrailingDotOrSeparator(absPath, localPath, os.PathSeparator), nil
}

func copyFromContainer(ctx context.Context, cli *Cli, srcContainer, srcPath, dstPath string) (err error) {
	apiClient := cli.Client()

	if dstPath != "-" {
		// Get an absolute destination path.
		dstPath, err = resolveLocalPath(dstPath)
		if err != nil {
			return err
		}
	}

	content, stat, err := apiClient.CopyFromContainer(ctx, srcContainer, srcPath)
	if err != nil {
		return err
	}
	defer content.Close()

	if dstPath == "-" {
		// Send the response to STDOUT.
		_, err = io.Copy(os.Stdout, content)
		return err
	}

	// Prepare source copy info.
	srcInfo := archive.CopyInfo{
		Path:       srcPath,
		Exists:     true,
		IsDir:      os.FileMode(stat.Mode).IsDir(),
		RebaseName: "",
	}

	preArchive := content
	if len(srcInfo.RebaseName) != 0 {
		_, srcBase := archive.SplitPathDirEntry(srcInfo.Path)
		preArchive = archive.RebaseArchiveEntries(content, srcBase, srcInfo.RebaseName)
	}
	// See comments in the implementation of `archive.CopyTo` for exactly what
	// goes into deciding how and whether the source archive needs to be
	// altered for the correct copy behavior.
	return archive.CopyTo(preArchive, srcInfo, dstPath)
}

func copyToContainer(ctx context.Context, cli *Cli, srcPath, dstContainer, dstPath string) (err error) {
	apiClient := cli.Client()

	if srcPath != "-" {
		// Get an absolute source path.
		srcPath, err = resolveLocalPath(srcPath)
		if err != nil {
			return err
		}
	}

	dstInfo := archive.CopyInfo{Path: dstPath}
	dstStat, err := apiClient.ContainerStatPath(ctx, dstContainer, dstPath)

	// Ignore any error and assume that the parent directory of the destination
	// path exists, in which case the copy may still succeed. If there is any
	// type of conflict (e.g., non-directory overwriting an existing directory
	// or vice versa) the extraction will fail. If the destination simply did
	// not exist, but the parent directory does, the extraction will still
	// succeed.
	if err == nil {
		dstInfo.Exists, dstInfo.IsDir = true, os.FileMode(dstStat.Mode).IsDir()
	}

	var (
		content         io.Reader
		resolvedDstPath string
	)

	if srcPath == "-" {
		// Use STDIN.
		content = os.Stdin
		resolvedDstPath = dstInfo.Path
		if !dstInfo.IsDir {
			return fmt.Errorf("destination %q must be a directory", fmt.Sprintf("%s:%s", dstContainer, dstPath))
		}
	} else {
		// Prepare source copy info.
		srcInfo, err := archive.CopyInfoSourcePath(srcPath, false)
		if err != nil {
			return err
		}

		srcArchive, err := archive.TarResource(srcInfo)
		if err != nil {
			return err
		}
		defer srcArchive.Close()

		dstDir, preparedArchive, err := archive.PrepareArchiveCopy(srcArchive, srcInfo, dstInfo)
		if err != nil {
			return err
		}
		defer preparedArchive.Close()

		resolvedDstPath = dstDir
		content = preparedArchive
	}

	return apiClient.CopyToContainer(ctx, dstContainer, resolvedDstPath, content)
}

// copyExample shows examples in copy command, and is used in auto-generated cli docs.
func copyExample() string {
	return `$ pouch cp 8assd1234:/root/foo /home
$ pouch cp /home/bar 712yasbc:/root`
}
