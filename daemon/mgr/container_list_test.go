package mgr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFilterContext(t *testing.T) {
	assert := assert.New(t)

	// test successful
	for _, t := range []*ContainerListOption{
		nil, {}, {Filter: map[string][]string{
			"label": {"a=b"},
		}},
	} {
		_, err := newFilterContext(t)
		assert.NoError(err)
	}

	// test failed
	for _, t := range []*ContainerListOption{
		{Filter: map[string][]string{
			"labels": {"a=b"},
		}},
		{Filter: map[string][]string{
			"foo": {},
		}},
	} {
		_, err := newFilterContext(t)
		assert.Error(err)
	}
}

func TestMatchFilter(t *testing.T) {
	assert := assert.New(t)
	type tCase struct {
		field    string
		value    string
		isFilter bool
	}

	option := ContainerListOption{
		Filter: map[string][]string{
			"id":     {"id1", "id2"},
			"status": {"running"},
			"name":   {"name3", "name4", "name5"},
		},
	}

	fc, err := newFilterContext(&option)
	assert.NoError(err)

	for _, t := range []tCase{
		{
			field:    "id",
			value:    "id1",
			isFilter: true,
		},
		{
			field:    "id",
			value:    "foor",
			isFilter: false,
		},
		{
			field:    "name",
			value:    "foor",
			isFilter: false,
		},
		{
			field:    "name",
			value:    "name3",
			isFilter: true,
		},
		{
			field:    "status",
			value:    "no",
			isFilter: false,
		},
		{
			field:    "status",
			value:    "running",
			isFilter: true,
		},
		{
			// since filter will return true when field not exist
			field:    "foo",
			value:    "bar",
			isFilter: true,
		},
	} {
		assert.Equal(t.isFilter, fc.matchFilter(t.field, t.value))
	}

}

func TestMatchKVFilter(t *testing.T) {
	assert := assert.New(t)
	type tCase struct {
		field    string
		value    map[string]string
		isFilter bool
	}

	option := ContainerListOption{
		Filter: map[string][]string{
			"label": {"foo", "hello=word"},
		},
	}

	fc, err := newFilterContext(&option)
	assert.NoError(err)

	for _, t := range []tCase{
		{
			field:    "label",
			value:    map[string]string{"a": "b"},
			isFilter: false,
		},
		{
			field:    "label",
			value:    map[string]string{"bar": "a"},
			isFilter: false,
		},
		{
			field:    "label",
			value:    map[string]string{"hello": "lala"},
			isFilter: false,
		},
		{
			field:    "label",
			value:    map[string]string{"foo": "a"},
			isFilter: true,
		},
		{
			field:    "label",
			value:    map[string]string{"hello": "word"},
			isFilter: true,
		},
	} {
		assert.Equal(t.isFilter, fc.matchKVFilter(t.field, t.value), fmt.Sprintf("%+v", t.value))
	}

	option = ContainerListOption{
		Filter: map[string][]string{
			"label": {"a!=b", "d!=c"},
		},
	}

	fc, err = newFilterContext(&option)
	assert.NoError(err)

	for _, t := range []tCase{
		{
			field:    "label",
			value:    map[string]string{"a": "b"},
			isFilter: false,
		},
		{
			field:    "label",
			value:    map[string]string{"bar": "a"},
			isFilter: false,
		},
		{
			field:    "label",
			value:    map[string]string{"a": "c"},
			isFilter: true,
		},
		{
			field:    "label",
			value:    map[string]string{"d": "c"},
			isFilter: false,
		},
		{
			field:    "label",
			value:    map[string]string{"d": "word"},
			isFilter: true,
		},
	} {
		assert.Equal(t.isFilter, fc.matchKVFilter(t.field, t.value), fmt.Sprintf("%+v", t.value))
	}
}
