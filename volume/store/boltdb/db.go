package boltdb

import (
	"os"
	"path"
	"time"

	"github.com/alibaba/pouch/volume/store"

	boltdb "github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

var defaultBucket = []byte("volume")

func init() {
	store.RegisterMetaStore(&bolt{})
}

type bolt struct {
	db *boltdb.DB
}

// New is used to make bolt metadata store instance.
func (b *bolt) New(metaPath string) error {
	opt := &boltdb.Options{
		Timeout: time.Second * 10,
	}

	dirname := path.Dir(metaPath)
	if _, err := os.Stat(dirname); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(dirname, 0755); err != nil {
			return errors.Wrap(err, "failed to create metadata path")
		}
	}

	db, err := boltdb.Open(metaPath, 0644, opt)
	if err != nil {
		return err
	}
	if err := b.prepare(db); err != nil {
		return err
	}
	b.db = db
	return nil
}

func (b *bolt) prepare(db *boltdb.DB) error {
	return db.Update(func(tx *boltdb.Tx) error {
		_, err := tx.CreateBucketIfNotExists(defaultBucket)
		if err != nil {
			return errors.Wrap(err, "boltdb: create bucket")
		}
		return nil
	})
}

// Put is used to put metadate into boltdb.
func (b *bolt) Put(key []byte, value []byte) error {
	return b.db.Update(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		if bucket == nil {
			return errors.New("boltdb: not found bucket")
		}
		if err := bucket.Put(key, value); err != nil {
			return errors.Wrap(err, "boltdb: put key: "+string(key))
		}
		return nil
	})
}

// Del is used to delete metadate from boltdb.
func (b *bolt) Del(key []byte) error {
	return b.db.Update(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		if bucket == nil {
			return errors.New("boltdb: not found bucket")
		}
		return bucket.Delete(key)
	})
}

// Get returns metadata from boltdb.
func (b *bolt) Get(key []byte) ([]byte, error) {
	var value []byte

	err := b.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		if bucket == nil {
			return errors.New("boltdb: not found bucket")
		}
		if value = bucket.Get(key); value == nil {
			return errors.New("boltdb: metadata not found")
		}
		return nil
	})

	return value, err
}

// List returns all metadata in boltdb.
func (b *bolt) List() ([][]byte, error) {
	values := make([][]byte, 0, 20)

	err := b.db.View(func(tx *boltdb.Tx) error {
		bucket := tx.Bucket(defaultBucket)
		if bucket == nil {
			return errors.New("boltdb: not found bucket")
		}

		return bucket.ForEach(func(k, v []byte) error {
			values = append(values, v)
			return nil
		})
	})

	return values, err
}
