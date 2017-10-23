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
