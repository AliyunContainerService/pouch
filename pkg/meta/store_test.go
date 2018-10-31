package meta

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/pkg/utils"
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

var (
	localBuckets = []Bucket{
		{MetaJSONFile, reflect.TypeOf(Demo{})},
		{"test.json", reflect.TypeOf(Demo2{})},
	}

	boltdbBuckets = []Bucket{
		{"boltdb", reflect.TypeOf(Demo3{})},
	}
)

func initStore(dbFile, driver string, buckets []Bucket) (*Store, error) {
	localCfg := Config{
		Driver:  driver,
		BaseDir: dbFile,
		Buckets: buckets,
	}

	return NewStore(localCfg)
}

func testStoreWrapper(t *testing.T, name, driver string, buckets []Bucket, test func(t *testing.T, s *Store)) {
	// initialize
	dbFile := path.Join("/tmp", utils.RandString(8, name, ""))
	if err := ensureFileNotExist(dbFile); err != nil {
		t.Fatal(err)
	}
	defer ensureFileNotExist(dbFile)

	s, err := initStore(dbFile, driver, buckets)
	if err != nil {
		t.Fatal(err)
	}

	test(t, s)
}

func ensureFileNotExist(file string) error {
	_, err := os.Stat(file)
	if err == nil {
		os.RemoveAll(file)
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	return nil
}

func testPut(t *testing.T, s *Store) {
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

func TestPut(t *testing.T) {
	testStoreWrapper(t, "TestPut", "local", localBuckets, testPut)
}

func testGet(t *testing.T, s *Store) {
	// put 1
	if err := s.Put(&Demo{
		A: 1,
		B: "key",
	}); err != nil {
		t.Fatal(err)
	}

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

func TestGet(t *testing.T) {
	testStoreWrapper(t, "TestGet", "local", localBuckets, testGet)
}

func testFetch(t *testing.T, s *Store) {
	// put 2
	if err := s.Put(&Demo{
		A: 2,
		B: "key2",
	}); err != nil {
		t.Fatal(err)
	}

	d := &Demo{}
	d.B = "key2"

	if err := s.Fetch(d); err != nil {
		t.Fatal(err)
	}

	if d.A != 2 || d.B != "key2" {
		t.Fatalf("not demo")
	}
}

func TestFetch(t *testing.T) {
	testStoreWrapper(t, "TestFetch", "local", localBuckets, testFetch)
}

func testList(t *testing.T, s *Store) {
	// put 2
	if err := s.Put(&Demo{
		A: 2,
		B: "key2",
	}); err != nil {
		t.Fatal(err)
	}

	objs, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(objs) != 1 {
		t.Fatalf("failed to list")
	}

}

func TestList(t *testing.T) {
	testStoreWrapper(t, "TestList", "local", localBuckets, testList)
}

func testRemove(t *testing.T, s *Store) {
	// put 1
	if err := s.Put(&Demo{
		A: 1,
		B: "key",
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.Remove("key"); err != nil {
		t.Fatal(err)
	}
}

func TestRemove(t *testing.T) {
	testStoreWrapper(t, "TestRemove", "local", localBuckets, testList)
}

type Demo3 struct {
	A int
	B string
}

func (d *Demo3) Key() string {
	return d.B
}

func testBoltdbPut(t *testing.T, boltdbStore *Store) {
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

func TestBoltdbPut(t *testing.T) {
	testStoreWrapper(t, "TestBoltdbPut", "boltdb", boltdbBuckets, testBoltdbPut)
}

func testBoltdbGet(t *testing.T, boltdbStore *Store) {
	// first put 1
	if err := boltdbStore.Put(&Demo3{
		A: 1,
		B: "key",
	}); err != nil {
		t.Fatal(err)
	}

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

func TestBoltdbGet(t *testing.T) {
	testStoreWrapper(t, "TestBoltdbGet", "boltdb", boltdbBuckets, testBoltdbGet)
}

func testBoltdbFetch(t *testing.T, boltdbStore *Store) {
	// first put 2
	if err := boltdbStore.Put(&Demo3{
		A: 2,
		B: "key2",
	}); err != nil {
		t.Fatal(err)
	}

	d := &Demo3{}
	d.B = "key2"

	if err := boltdbStore.Fetch(d); err != nil {
		t.Fatal(err)
	}

	if d.A != 2 || d.B != "key2" {
		t.Fatalf("not demo")
	}
}

func TestBoltdbFetch(t *testing.T) {
	testStoreWrapper(t, "TestBoltdbFetch", "boltdb", boltdbBuckets, testBoltdbFetch)
}

func testBoltdbList(t *testing.T, boltdbStore *Store) {
	// first put 2
	if err := boltdbStore.Put(&Demo3{
		A: 2,
		B: "key2",
	}); err != nil {
		t.Fatal(err)
	}

	objs, err := boltdbStore.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(objs) != 1 {
		t.Fatalf("failed to list")
	}
}

func TestBoltdbList(t *testing.T) {
	testStoreWrapper(t, "TestBoltdbList", "boltdb", boltdbBuckets, testBoltdbList)
}

func testBoltdbRemove(t *testing.T, boltdbStore *Store) {
	// first put 1
	if err := boltdbStore.Put(&Demo3{
		A: 1,
		B: "key",
	}); err != nil {
		t.Fatal(err)
	}

	if err := boltdbStore.Remove("key"); err != nil {
		t.Fatal(err)
	}
}

func TestBoltdbRemove(t *testing.T) {
	testStoreWrapper(t, "TestBoltdbRemove", "boltdb", boltdbBuckets, testBoltdbRemove)
}

func testBoltdbClose(t *testing.T, boltdbStore *Store) {
	if err := boltdbStore.Shutdown(); err != nil {
		t.Fatal(err)
	}

	// test List again, should occur error here
	_, err := boltdbStore.List()
	if err == nil {
		t.Fatal(fmt.Errorf("still can visit the boltdb after execute close db action"))
	}
}

func TestBoltdbClose(t *testing.T) {
	testStoreWrapper(t, "TestBoltdbClose", "boltdb", boltdbBuckets, testBoltdbClose)
}

type Demo4 struct {
	A int
	B string
}

func (d *Demo4) Key() string {
	return d.B
}

func testKeysWithPrefix(t *testing.T, boltdb4 *Store) {
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

func TestKeysWithPrefix(t *testing.T) {
	testStoreWrapper(t, "TestKeysWithPrefix", "boltdb", boltdbBuckets, testKeysWithPrefix)
}
