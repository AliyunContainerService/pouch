package containerio

import (
	"reflect"
	"testing"

	"github.com/alibaba/pouch/pkg/collect"
)

var testCache = NewCache()

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

	var f fields
	f.m = testCache.m

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "cachePutIONil",
			fields: f,
			args: args{
				id: "123",
				io: nil,
			},
			wantErr: false,
		},
		{
			name:   "cachePutIdNil",
			fields: f,
			args: args{
				id: "",
				io: nil,
			},
			wantErr: false,
		},
		{
			name:   "cachePutIdDup",
			fields: f,
			args: args{
				id: "123",
				io: &IO{},
			},
			wantErr: false,
		},
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
		{
			name: "cacheGetNilId",
			fields: fields{
				m: testCache.m,
			},
			args: args{
				id: "",
			},
			want: nil,
		},
		{
			name: "cacheGetIdNotFound",
			fields: fields{
				m: testCache.m,
			},
			args: args{
				id: "abc",
			},
			want: nil,
		},
		{
			name: "cacheGetIdNotFound",
			fields: fields{
				m: testCache.m,
			},
			args: args{
				id: "123",
			},
			want: &IO{},
		},
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
		{
			name: "cacheRmOk",
			fields: fields{
				m: testCache.m,
			},
			args: args{
				id: "123",
			},
		},
		{
			name: "cacheRmNil",
			fields: fields{
				m: testCache.m,
			},
			args: args{
				id: "",
			},
		},
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
