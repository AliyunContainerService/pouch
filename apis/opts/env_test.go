package opts

import (
	"fmt"
	"testing"
        "strings"
        "reflect"
)

func ParseEnv(env []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, e := range env {
		fields := strings.SplitN(e, "=", 2)
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid env %s: env must be in format of key=value", e)
		}
		results[fields[0]] = fields[1]
	}

	return results, nil
}

func TestParseEnv(t *testing.T) {
	type args struct {
		env []string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{env: []string{"foo=bar"}}, want: map[string]string{"foo": "bar"}, wantErr: false},
		{name: "test2", args: args{env: []string{"ErrorInfo"}}, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEnv(tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
