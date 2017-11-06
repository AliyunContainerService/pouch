package store

import (
	"strings"

	"github.com/alibaba/pouch/pkg/serializer"
	volerr "github.com/alibaba/pouch/volume/error"
	"github.com/alibaba/pouch/volume/types"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var ms metaStorer

// RegisterMetaStore is used to register a metadata store.
func RegisterMetaStore(m metaStorer) {
	ms = m
}

type metaStorer interface {
	New(path string) error
	Put(key []byte, value []byte) error
	Del(key []byte) error
	Get(key []byte) ([]byte, error)
	List() ([][]byte, error)
}

// MetaNewStore is used to make a metadata store instance.
func MetaNewStore(path string) error {
	return ms.New(path)
}

// MetaPut is used to put volume metadate into store.
func MetaPut(v *types.Volume) error {
	value, err := serializer.Codec.Encode(v)
	if err != nil {
		return errors.Wrap(err, "meta: encode volume")
	}
	return ms.Put([]byte(v.GetName()), value)
}

// MetaDel is used to delete a volume metadata by name.
func MetaDel(name string) error {
	return ms.Del([]byte(name))
}

// MetaGet returns a volume metadata.
func MetaGet(v *types.Volume) error {
	value, err := ms.Get([]byte(v.GetName()))
	if err != nil {
		if strings.Contains(err.Error(), "metadata not found") {
			return volerr.ErrLocalMetaNotfound
		}
		return err
	}

	if err := serializer.Codec.Decode(value, v); err != nil {
		return errors.Wrap(err, "meta: decode volume")
	}
	return nil
}

// MetaList returns all volumes metadata.
func MetaList() ([]*types.Volume, error) {
	values, err := ms.List()
	if err != nil {
		return nil, err
	}

	volumes := make([]*types.Volume, 0, len(values))

	for _, value := range values {
		v := &types.Volume{}
		if err := serializer.Codec.Decode(value, v); err != nil {
			log.Errorf("meta list: decode error: %v", err)
			continue
		}
		volumes = append(volumes, v)
	}

	return volumes, nil
}
