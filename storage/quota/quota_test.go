// +build linux

package quota

import "testing"

func TestGetDefaultQuota(t *testing.T) {
	type args struct {
		quotas map[string]string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal case with supposed data root /",
			args: args{
				quotas: map[string]string{
					"/": "1000kb",
				},
			},
			want: "1000kb",
		},
		{
			name: "normal case with supposed data .*",
			args: args{
				quotas: map[string]string{
					".*": "2000kb",
				},
			},
			want: "2000kb",
		},
		{
			name: "normal case with supposed data .* and /",
			args: args{
				quotas: map[string]string{
					".*": "2000kb",
					"/":  "1000kb",
				},
			},
			want: "1000kb",
		},
		{
			name: "normal case with no supposed data",
			args: args{
				quotas: map[string]string{
					"asdfghj": "2000kb",
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetDefaultQuota(tt.args.quotas); got != tt.want {
				t.Errorf("GetDefaultQuota() = %v, want %v", got, tt.want)
			}
		})
	}
}
