package main

import (
	"context"
	"fmt"

	"github.com/alibaba/pouch/pkg/reference"

	"github.com/spf13/cobra"
)

// pushDescription is used to describe push command in detail and auto generate command doc.
var pushDescription = "Push a local image to remote registry."

// PushCommand is used to implement 'push' command, it pushes image to some registries.
type PushCommand struct {
	baseCommand
}

// Init initializes push command.
func (p *PushCommand) Init(c *Cli) {
	p.cli = c

	p.cmd = &cobra.Command{
		Use:   "push IMAGE[:TAG]",
		Short: "Push an image to registry",
		Args:  cobra.ExactArgs(1),
		Long:  pushDescription,
		RunE: func(_ *cobra.Command, args []string) error {
			return p.runPush(args[0])
		},
		Example: p.pushExample(),
	}
}

// runPush pushes a image.
func (p *PushCommand) runPush(refImage string) error {
	apiClient := p.cli.Client()

	namedRef, err := reference.Parse(refImage)
	if err != nil {
		return err
	}
	namedRef = reference.TrimTagForDigest(reference.WithDefaultTagIfMissing(namedRef))

	responseBody, err := apiClient.ImagePush(context.TODO(), namedRef.String(), fetchRegistryAuth(namedRef.Name()))
	if err != nil {
		return fmt.Errorf("failed to push image: %v", err)
	}
	defer responseBody.Close()

	return showProgress(responseBody)
}

// pushExample shows examples in push command, and is used in auto-generated cli docs.
func (p *PushCommand) pushExample() string {
	return `$ pouch push docker.io/testing/busybox:1.25
docker.io/testing/busybox:1.25:                                                   resolved |++++++++++++++++++++++++++++++++++++++|
manifest-sha256:29f5d56d12684887bdfa50dcd29fc31eea4aaf4ad3bec43daf19026a7ce69912: done
layer-sha256:56bec22e355981d8ba0878c6c2f23b21f422f30ab0aba188b54f1ffeff59c190:    done
config-sha256:e02e811dd08fd49e7f6032625495118e63f597eb150403d02e3238af1df240ba:   done
elapsed: 0.0 s                                                                    total:   0.0 B (0.0 B/s)
`
}
