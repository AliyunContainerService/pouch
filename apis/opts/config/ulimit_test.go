package config

import (
	"testing"

	"github.com/docker/go-units"
)

func TestUlimit_Set(t *testing.T) {
	type fields struct {
		values map[string]*units.Ulimit
	}
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "invalid args: too many limit value arguments",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "core=3:4:5",
			},
			wantErr: true,
		},
		{
			name: "invalid args: inexistent ulimit type",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "inexistent=3:4",
			},
			wantErr: true,
		},
		{
			name: "invalid args: ulimit soft limit is greater than hard limit",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "core=4:3",
			},
			wantErr: true,
		},
		{
			name: "valid args and ulimit is nil",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "core=3:4",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Ulimit{
				values: tt.fields.values,
			}
			if err := u.Set(tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("Ulimit.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
