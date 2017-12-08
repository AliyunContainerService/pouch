package mgr

import (
	"fmt"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/meta"
	"github.com/alibaba/pouch/pkg/errtypes"
	"github.com/alibaba/pouch/pkg/randomid"

	"github.com/pkg/errors"
)

// containerID returns the container's id, the parameter 'nameOrPrefix' may be container's
// name, id or prefix id.
func (mgr *ContainerManager) containerID(nameOrPrefix string) (string, error) {
	var obj meta.Object

	// name is the container's name.
	id, ok := mgr.NameToID.Get(nameOrPrefix).String()
	if ok {
		return id, nil
	}

	// name is the container's prefix of the id.
	objs, err := mgr.Store.GetWithPrefix(nameOrPrefix)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get container info with prefix: %s", nameOrPrefix)
	}
	if len(objs) > 1 {
		return "", errors.Wrap(errtypes.ErrTooMany, "container: "+nameOrPrefix)
	} else if len(objs) == 0 {
		return "", errors.Wrap(errtypes.ErrNotfound, "container: "+nameOrPrefix)
	}
	obj = objs[0]

	meta, ok := obj.(*types.ContainerInfo)
	if !ok {
		return "", fmt.Errorf("failed to get container info, invalid meta's type")
	}

	return meta.ID, nil
}

func (mgr *ContainerManager) container(nameOrPrefix string) (*Container, error) {
	res, ok := mgr.cache.Get(nameOrPrefix).Result()
	if ok {
		return res.(*Container), nil
	}

	id, err := mgr.containerID(nameOrPrefix)
	if err != nil {
		return nil, err
	}

	// lookup again
	res, ok = mgr.cache.Get(id).Result()
	if ok {
		return res.(*Container), nil
	}

	return nil, errors.Wrap(errtypes.ErrNotfound, "container "+nameOrPrefix)
}

// generateID generates an ID for newly created container. We must ensure that
// this ID has not used yet.
func (mgr *ContainerManager) generateID() (string, error) {
	var id string
	for {
		id = randomid.Generate()
		_, err := mgr.Store.Get(id)
		if err != nil {
			if merr, ok := err.(meta.Error); ok && merr.IsNotfound() {
				break
			}
			return "", err
		}
	}
	return id, nil
}

// generateName generates container name by container ID.
// It get first 6 continuous letters which has not been taken.
// TODO: take a shorter than 6 letters ID into consideration.
// FIXME: there is possibility that for block loops forever.
func (mgr *ContainerManager) generateName(id string) string {
	var name string
	i := 0
	for {
		if i+6 > len(id) {
			break
		}
		name = id[i : i+6]
		i++
		if !mgr.NameToID.Get(name).Exist() {
			break
		}
	}
	return name
}
