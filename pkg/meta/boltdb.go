package meta

import (
	"os"
	"path"
	"time"

	boltdb "github.com/boltdb/bolt"
	"github.com/pkg/errors"
)

func init() {
	Register("boltdb", &bolt{})
}

type bolt struct {
	db *boltdb.DB
}

// New is used to make bolt metadata store instance.
func (b *bolt) New(cfg Config) error {
	opt := &boltdb.Options{
		Timeout: time.Second * 10,
	}

	if _, err := os.Stat(cfg.BaseDir); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(path.Dir(cfg.BaseDir), 0755); err != nil {
			return errors.Wrap(err, "failed to create metadata path")
		}
	}

	db, err := boltdb.Open(cfg.BaseDir, 0644, opt)
	if err != nil {
		return err
	}
	for _, bucket := range cfg.Buckets {
		if err := b.prepare(db, []byte(bucket.Name)); err != nil {
			return err
		}
	}
	b.db = db

	return nil
}

func (b *bolt) prepare(db *boltdb.DB, bucket []byte) error {
	return db.Update(func(tx *boltdb.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return errors.Wrap(err, "failed to create bucket in boltdb")
		}
		return nil
	})
}

// Path returns boltdb store file.
func (b *bolt) Path(key string) string {
	return b.db.Path()
}

// Keys return all keys for boltdb.
func (b *bolt) Keys(bucket string) ([]string, error) {
	keys := make([]string, 0)

	err := b.db.View(func(tx *boltdb.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return ErrBucketNotFound
		}

		return bkt.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})

	return keys, err
}

// Put is used to put metadate into boltdb.
func (b *bolt) Put(bucket, key string, value []byte) error {
	return b.db.Update(func(tx *boltdb.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return ErrBucketNotFound
		}
		if err := bkt.Put([]byte(key), value); err != nil {
			return errors.Wrapf(err, "failed to put key %s in boltdb", key)
		}
		return nil
	})
}

// Del is used to delete metadate from boltdb.
func (b *bolt) Remove(bucket string, key string) error {
	return b.db.Update(func(tx *boltdb.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return ErrBucketNotFound
		}
		return bkt.Delete([]byte(key))
	})
}

// Get returns metadata from boltdb.
func (b *bolt) Get(bucket string, key string) ([]byte, error) {
	var value []byte

	err := b.db.View(func(tx *boltdb.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return ErrBucketNotFound
		}
		if value = bkt.Get([]byte(key)); value == nil {
			return ErrObjectNotFound
		}
		return nil
	})

	return value, err
}

// List returns all metadata in boltdb.
func (b *bolt) List(bucket string) ([][]byte, error) {
	values := make([][]byte, 0, 20)

	err := b.db.View(func(tx *boltdb.Tx) error {
		bkt := tx.Bucket([]byte(bucket))
		if bkt == nil {
			return ErrBucketNotFound
		}

		return bkt.ForEach(func(k, v []byte) error {
			values = append(values, v)
			return nil
		})
	})

	return values, err
}
