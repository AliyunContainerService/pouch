package containerio

import (
	"reflect"
	"testing"

	"github.com/alibaba/pouch/pkg/collect"
)

func TestNewCache(t *testing.T) {
	tests := []struct {
		name string
		want *Cache
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCache(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCache() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache_Put(t *testing.T) {
	type fields struct {
		m *collect.SafeMap
	}
	type args struct {
		id string
		io *IO
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{
				m: tt.fields.m,
			}
			if err := c.Put(tt.args.id, tt.args.io); (err != nil) != tt.wantErr {
				t.Errorf("Cache.Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCache_Get(t *testing.T) {
	type fields struct {
		m *collect.SafeMap
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *IO
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{
				m: tt.fields.m,
			}
			if got := c.Get(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Cache.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCache_Remove(t *testing.T) {
	type fields struct {
		m *collect.SafeMap
	}
	type args struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cache{
				m: tt.fields.m,
			}
			c.Remove(tt.args.id)
		})
	}
}
