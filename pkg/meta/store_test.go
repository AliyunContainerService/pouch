package meta

import (
	"reflect"
	"testing"
)

type Demo struct {
	A int
	B string
}

func (d *Demo) Key() string {
	return d.B
}

type Demo2 struct {
	A int
	B string
}

func (d *Demo2) Key() string {
	return d.B
}

var s *Store

var cfg = Config{
	Driver:  "local",
	BaseDir: "/tmp/containers",
	Buckets: []Bucket{
		{MetaJSONFile, reflect.TypeOf(Demo{})},
		{"test.json", reflect.TypeOf(Demo2{})},
	},
}

func TestPut(t *testing.T) {
	// initialize
	var err error
	s, err = NewStore(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// put 1
	if err := s.Put(&Demo{
		A: 1,
		B: "key",
	}); err != nil {
		t.Fatal(err)
	}
	// put 2
	if err := s.Put(&Demo{
		A: 2,
		B: "key2",
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.Bucket("test.json").Put(&Demo2{
		A: 1,
		B: "key",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestGet(t *testing.T) {
	obj, err := s.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	if d, ok := obj.(*Demo); !ok {
		t.Fatalf("failed to get")
	} else {
		if d.A != 1 || d.B != "key" {
			t.Fatalf("not demo")
		}
	}
}

func TestFetch(t *testing.T) {
	d := &Demo{}
	d.B = "key2"

	if err := s.Fetch(d); err != nil {
		t.Fatal(err)
	}

	if d.A != 2 || d.B != "key2" {
		t.Fatalf("not demo")
	}
}

func TestList(t *testing.T) {
	objs, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(objs) != 2 {
		t.Fatalf("failed to list")
	}
}

func TestRemove(t *testing.T) {
	if err := s.Remove("key"); err != nil {
		t.Fatal(err)
	}
}

type Demo3 struct {
	A int
	B string
}

func (d *Demo3) Key() string {
	return d.B
}

var boltdbStore *Store

var boltdbCfg = Config{
	Driver:  "boltdb",
	BaseDir: "/tmp/bolt.db",
	Buckets: []Bucket{
		{"boltdb", reflect.TypeOf(Demo3{})},
	},
}

func TestBoltdbPut(t *testing.T) {
	// initialize
	var err error
	boltdbStore, err = NewStore(boltdbCfg)
	if err != nil {
		t.Fatal(err)
	}

	// put 1
	if err := boltdbStore.Put(&Demo3{
		A: 1,
		B: "key",
	}); err != nil {
		t.Fatal(err)
	}
	// put 2
	if err := boltdbStore.Put(&Demo3{
		A: 2,
		B: "key2",
	}); err != nil {
		t.Fatal(err)
	}

	if err := boltdbStore.Bucket("boltdb").Put(&Demo3{
		A: 1,
		B: "key",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestBoltdbGet(t *testing.T) {
	obj, err := boltdbStore.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	if d, ok := obj.(*Demo3); !ok {
		t.Fatalf("failed to get")
	} else {
		if d.A != 1 || d.B != "key" {
			t.Fatalf("not demo")
		}
	}
}

func TestBoltdbFetch(t *testing.T) {
	d := &Demo3{}
	d.B = "key2"

	if err := boltdbStore.Fetch(d); err != nil {
		t.Fatal(err)
	}

	if d.A != 2 || d.B != "key2" {
		t.Fatalf("not demo")
	}
}

func TestBoltdbList(t *testing.T) {
	objs, err := boltdbStore.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(objs) != 2 {
		t.Fatalf("failed to list")
	}
}

func TestBoltdbRemove(t *testing.T) {
	if err := boltdbStore.Remove("key"); err != nil {
		t.Fatal(err)
	}
}

type Demo4 struct {
	A int
	B string
}

func (d *Demo4) Key() string {
	return d.B
}

var boltdb4 *Store

var boltdbCfg4 = Config{
	Driver:  "boltdb",
	BaseDir: "/tmp/bolt4.db",
	Buckets: []Bucket{
		{"boltdb", reflect.TypeOf(Demo4{})},
	},
}

func TestKeysWithPrefix(t *testing.T) {
	var err error
	boltdb4, err = NewStore(boltdbCfg4)
	if err != nil {
		t.Fatal(err)
	}

	// put 1
	if err := boltdb4.Put(&Demo4{
		A: 1,
		B: "prefixkey",
	}); err != nil {
		t.Fatal(err)
	}
	// put 2
	if err := boltdb4.Put(&Demo4{
		A: 2,
		B: "prefixkey2",
	}); err != nil {
		t.Fatal(err)
	}

	// find item with prefix
	obj, err := boltdb4.KeysWithPrefix("prefixkey")
	if err != nil {
		t.Fatal(err)
	}
	if len(obj) != 2 {
		t.Fatal("should get 2 item")
	}

	// find item with prefix empty should return null item
	obj, err = boltdb4.KeysWithPrefix("")
	if err != nil {
		t.Fatal(err)
	}
	if len(obj) != 0 {
		t.Fatal("should get empty item")
	}
}
