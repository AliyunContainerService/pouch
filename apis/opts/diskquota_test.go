package opts

import (
	"fmt"
	"reflect"
	"testing"
)

func TestParseDiskQuota(t *testing.T) {
	type args struct {
		diskquota []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{diskquota: []string{""}}, want: nil, wantErr: true},
		{name: "test2", args: args{diskquota: []string{"foo=foo=foo"}}, want: nil, wantErr: true},
		{name: "test3", args: args{diskquota: []string{"foo"}}, want: map[string]string{".*": "foo"}, wantErr: false},
		{name: "test4", args: args{diskquota: []string{"foo=foo"}}, want: map[string]string{"foo": "foo"}, wantErr: false},
		{name: "test5", args: args{diskquota: []string{"foo=foo", "bar=bar"}}, want: map[string]string{"foo": "foo", "bar": "bar"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDiskQuota(tt.args.diskquota)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDiskQuota() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDiskQuota() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseQuotaID(t *testing.T) {
	tests := []struct {
		id        string
		quota     []string
		expectID  string
		expectErr error
	}{
		{id: "", quota: []string{}, expectID: "", expectErr: nil},
		{id: "", quota: []string{"20G"}, expectID: "-1", expectErr: nil},
		{id: "", quota: []string{"20G", "/abc=10G"}, expectID: "", expectErr: nil},
		{id: "0", quota: []string{}, expectID: "0", expectErr: nil},
		{id: "0", quota: []string{"20G"}, expectID: "-1", expectErr: nil},
		{id: "0", quota: []string{"20G", "/abc=10G"}, expectID: "0", expectErr: nil},
		{id: "-1", quota: []string{}, expectID: "", expectErr: fmt.Errorf("invalid to set quota id(-1) without disk-quota")},
		{id: "-1", quota: []string{"20G"}, expectID: "-1", expectErr: nil},
		{id: "-1", quota: []string{"20G", "/abc=10G"}, expectID: "", expectErr: fmt.Errorf("invalid to set quota id(-1) for multi disk-quota")},
		{id: "1", quota: []string{}, expectID: "", expectErr: fmt.Errorf("invalid to set quota id(1) without disk-quota")},
		{id: "1", quota: []string{"20G"}, expectID: "1", expectErr: nil},
		{id: "1", quota: []string{"20G", "/abc=10G"}, expectID: "", expectErr: fmt.Errorf("invalid to set quota id(1) for multi disk-quota")},
	}
	for _, tt := range tests {
		got, err := ParseQuotaID(tt.id, tt.quota)
		if (err != nil && tt.expectErr == nil) ||
			(err == nil && tt.expectErr != nil) {
			t.Fatalf("ParseQuotaID(%v %v) error = %v, wantErr %v", tt.id, tt.quota, err, tt.expectErr)
		}
		if err != nil && tt.expectErr != nil && err.Error() != tt.expectErr.Error() {
			t.Fatalf("ParseQuotaID(%v %v) error = %v, wantErr %v", tt.id, tt.quota, err, tt.expectErr)
		}
		if got != tt.expectID {
			t.Fatalf("ParseQuotaID(%v %v) = %v, want %v", tt.id, tt.quota, got, tt.expectID)
		}

	}
}
