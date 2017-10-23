package ctrd

import (
	"context"
	"testing"
	"time"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func TestPullImage(t *testing.T) {
	cli, err := NewClient(Config{})
	if err != nil {
		t.Fatal(err)
	}

	h := func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		println(desc.Digest, desc.Size)
		return nil, nil
	}

	err = cli.PullImage(context.TODO(), "docker.io/library/redis:alpine", h)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateContainer(t *testing.T) {
	cli, err := NewClient(Config{})
	if err != nil {
		t.Fatal(err)
	}

	id := "12345678abc"
	ref := "docker.io/library/redis:alpine"
	_ = ref

	err = cli.RecoverContainer(context.TODO(), id)
	if err != nil {
		t.Error(err)
	}
	_, err = cli.DestroyContainer(context.TODO(), id)
	if err != nil {
		t.Error(err)
	}

	// for i := 0; i < 1; i++ {
	// 	err = cli.createContainer(context.TODO(), ref, id, nil)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}

	// 	time.Sleep(3 * time.Second)

	// 	_, err = cli.DestroyContainer(context.TODO(), id)
	// 	if err != nil {
	// 		t.Fatal(err)
	// 	}
	// }

	err = cli.createContainer(context.TODO(), ref, id, nil)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	all, err := cli.listContainerStore(context.TODO())
	if err == nil {
		for i, a := range all {
			t.Logf("%d: %s", i, a)
		}
	}
}
