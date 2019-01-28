package collect

import (
	"testing"

	"github.com/alibaba/pouch/pkg/utils"

	"github.com/stretchr/testify/assert"
)

// TestSafeMapValues is to valid the Values function
func TestSafeMapValues(t *testing.T) {
	safeMap := NewSafeMap()
	assert.Equal(t, len(safeMap.inner), 0)

	safeMap.Put("key", "value")
	assert.Equal(t, len(safeMap.inner), 1)

	safeMap.Put("key1", "value1")
	assert.Equal(t, len(safeMap.inner), 2)

	// first not specify filter
	values := safeMap.Values(nil)
	assert.Equal(t, len(values), 2)
	for k, v := range values {
		if !utils.StringInSlice([]string{"key", "key1"}, k) {
			t.Errorf("got unexpected key: %v", k)
		}

		if k == "key" && v != "value" {
			t.Errorf("expected 'value', but got %v", v)
		}

		if k == "key1" && v != "value1" {
			t.Errorf("expected 'value1', but got %v", v)
		}
	}

	// test filter function
	fn := func(obj interface{}) bool {
		v, ok := obj.(string)
		if !ok || v != "value" {
			return false
		}

		return true
	}

	values = safeMap.Values(fn)
	assert.Equal(t, len(values), 1)
	v, ok := values["key"]
	if !ok {
		t.Errorf("expected 'key' in map, but got: %v", values)
	}

	stringVar, ok := v.(string)
	if !ok || stringVar != "value" {
		t.Errorf("expected map's value is 'value', but got: %v", stringVar)
	}
}

func TestSafeMapPutAndGet(t *testing.T) {
	safeMap := NewSafeMap()
	assert.Equal(t, len(safeMap.inner), 0)

	safeMap.Put("key", "value")
	assert.Equal(t, len(safeMap.inner), 1)

	value := safeMap.Get("key")
	assert.Equal(t, len(safeMap.inner), 1)
	assert.Equal(t, value.ok, true)
	assert.Equal(t, value.data, "value")

	// there is no key named non-exist
	value = safeMap.Get("non-exist")
	assert.Equal(t, len(safeMap.inner), 1)
	assert.Equal(t, value.ok, false)
	assert.Equal(t, value.data, nil)

	// get key twice
	value = safeMap.Get("key")
	assert.Equal(t, len(safeMap.inner), 1)
	assert.Equal(t, value.ok, true)
	assert.Equal(t, value.data, "value")

	// put the same key with a new value
	safeMap.Put("key", "value2")
	assert.Equal(t, len(safeMap.inner), 1)

	// get key twice, and be supposed to get new value
	value = safeMap.Get("key")
	assert.Equal(t, len(safeMap.inner), 1)
	assert.Equal(t, value.ok, true)
	assert.Equal(t, value.data, "value2")

	// put new keys with new value
	safeMap.Put("key2", []string{"asdfgh", "123344"})
	assert.Equal(t, len(safeMap.inner), 2)

	value = safeMap.Get("key2")
	assert.Equal(t, len(safeMap.inner), 2)
	assert.Equal(t, value.ok, true)
	assert.Equal(t, value.data, []string{"asdfgh", "123344"})
}

// TestSafeMapDirectNew test functions should not panic.
func TestSafeMapDirectNew(t *testing.T) {
	assert := assert.New(t)
	defer func() {
		if err := recover(); err != nil {
			t.Fatal(err)
		}
	}()

	sm := &SafeMap{}
	// test Put not panic
	sm.Put("k", "v")

	// test Remove not panic
	sm.Remove("k")

	// test Values not panic
	values := sm.Values(nil)
	assert.Equal(values, map[string]interface{}{})
}

func TestSafeMapRemove(t *testing.T) {
	safeMap := NewSafeMap()
	assert.Equal(t, len(safeMap.inner), 0)
	// remove a non-existence key
	safeMap.Remove("key")
	assert.Equal(t, len(safeMap.inner), 0)

	safeMap.Put("key", "value")
	assert.Equal(t, len(safeMap.inner), 1)

	safeMap.Remove("key")
	assert.Equal(t, len(safeMap.inner), 0)
}

func TestResult(t *testing.T) {
	value := &Value{
		data: "asdf",
		ok:   true,
	}

	data, ok := value.Result()
	assert.Equal(t, data, "asdf")
	assert.Equal(t, ok, true)
}

func TestExist(t *testing.T) {
	testCases := []*Value{
		{
			data: "asdf",
			ok:   true,
		},
		{
			data: []string{"asd"},
			ok:   true,
		},
		{
			data: nil,
			ok:   false,
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, testCase.Exist(), testCase.ok)
	}
}

func TestString(t *testing.T) {
	type Result struct {
		str string
		ok  bool
	}
	testCases := []struct {
		value  *Value
		result Result
	}{
		{
			value: &Value{
				data: "asdf",
				ok:   true,
			},
			result: Result{
				str: "asdf",
				ok:  true,
			},
		},
		{
			value: &Value{
				data: []string{"asdf"},
				ok:   true,
			},
			result: Result{
				str: "",
				ok:  false,
			},
		},
		{
			value: &Value{
				data: 11,
				ok:   true,
			},
			result: Result{
				str: "",
				ok:  false,
			},
		},
		{
			value: &Value{
				data: nil,
				ok:   false,
			},
			result: Result{
				str: "",
				ok:  false,
			},
		},
	}

	for _, testCase := range testCases {
		str, ok := testCase.value.String()
		assert.Equal(t, str, testCase.result.str)
		assert.Equal(t, ok, testCase.result.ok)
	}
}

func TestInt(t *testing.T) {
	type Result struct {
		result int
		ok     bool
	}
	testCases := []struct {
		value  *Value
		result Result
	}{
		{
			value: &Value{
				data: "asdf",
				ok:   true,
			},
			result: Result{
				result: 0,
				ok:     false,
			},
		},
		{
			value: &Value{
				data: []string{"asdf"},
				ok:   true,
			},
			result: Result{
				result: 0,
				ok:     false,
			},
		},
		{
			value: &Value{
				data: int(11),
				ok:   true,
			},
			result: Result{
				result: int(11),
				ok:     true,
			},
		},
		{
			value: &Value{
				data: nil,
				ok:   false,
			},
			result: Result{
				result: 0,
				ok:     false,
			},
		},
	}

	for _, testCase := range testCases {
		result, ok := testCase.value.Int()
		assert.Equal(t, result, testCase.result.result)
		assert.Equal(t, ok, testCase.result.ok)
	}
}

func TestInt32(t *testing.T) {
	type Result struct {
		result int32
		ok     bool
	}
	testCases := []struct {
		value  *Value
		result Result
	}{
		{
			value: &Value{
				data: "asdf",
				ok:   true,
			},
			result: Result{
				result: 0,
				ok:     false,
			},
		},
		{
			value: &Value{
				data: []string{"asdf"},
				ok:   true,
			},
			result: Result{
				result: 0,
				ok:     false,
			},
		},
		{
			value: &Value{
				data: int32(11),
				ok:   true,
			},
			result: Result{
				result: int32(11),
				ok:     true,
			},
		},
		{
			value: &Value{
				data: nil,
				ok:   false,
			},
			result: Result{
				result: 0,
				ok:     false,
			},
		},
	}

	for _, testCase := range testCases {
		result, ok := testCase.value.Int32()
		assert.Equal(t, result, testCase.result.result)
		assert.Equal(t, ok, testCase.result.ok)
	}
}

func TestInt64(t *testing.T) {
	type Result struct {
		result int64
		ok     bool
	}
	testCases := []struct {
		value  *Value
		result Result
	}{
		{
			value: &Value{
				data: "asdf",
				ok:   true,
			},
			result: Result{
				result: 0,
				ok:     false,
			},
		},
		{
			value: &Value{
				data: []string{"asdf"},
				ok:   true,
			},
			result: Result{
				result: 0,
				ok:     false,
			},
		},
		{
			value: &Value{
				data: int64(11),
				ok:   true,
			},
			result: Result{
				result: int64(11),
				ok:     true,
			},
		},
		{
			value: &Value{
				data: nil,
				ok:   false,
			},
			result: Result{
				result: 0,
				ok:     false,
			},
		},
	}

	for _, testCase := range testCases {
		result, ok := testCase.value.Int64()
		assert.Equal(t, result, testCase.result.result)
		assert.Equal(t, ok, testCase.result.ok)
	}
}
