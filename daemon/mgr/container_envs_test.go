package mgr

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Test_updateContainerEnv(t *testing.T) {
	type args struct {
		inputRawEnv []string
		baseFs      string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := updateContainerEnv(tt.args.inputRawEnv, tt.args.baseFs); (err != nil) != tt.wantErr {
				t.Errorf("updateContainerEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_combineLocalAndInputEnv(t *testing.T) {
	type args struct {
		inputEnv map[string]string
		localEnv map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "normal test case",
			args: args{
				inputEnv: map[string]string{
					"A": "B",
					"C": "D",
				},
				localEnv: map[string]string{
					"AA": "BB",
					"CC": "DD",
				},
			},
			want: map[string]string{
				"A":  "B",
				"C":  "D",
				"AA": "BB",
				"CC": "DD",
			},
		},
		{
			name: "normal test case with input env exists in local env, so to replace",
			args: args{
				inputEnv: map[string]string{
					"A": "B",
					"B": "C",
					"C": "D",
				},
				localEnv: map[string]string{
					"A": "C",
					"B": "C",
				},
			},
			want: map[string]string{
				"A": "B",
				"B": "C",
				"C": "D",
			},
		},
		{
			name: "normal test case with PATH both in input and local",
			args: args{
				inputEnv: map[string]string{
					"PATH": "/sbin",
					"key":  "value",
				},
				localEnv: map[string]string{
					"PATH": "/usr/local/bin:$PATH",
					"A":    "B",
				},
			},
			want: map[string]string{
				"PATH": "/sbin:$PATH",
				"key":  "value",
				"A":    "B",
			},
		},
		{
			name: "normal test case with no PATH in input",
			args: args{
				inputEnv: map[string]string{
					"key": "value",
				},
				localEnv: map[string]string{
					"PATH": "/usr/local/bin:$PATH",
					"A":    "B",
				},
			},
			want: map[string]string{
				"PATH": "/usr/local/bin:$PATH",
				"key":  "value",
				"A":    "B",
			},
		},
		{
			name: "normal test case with no PATH in local",
			args: args{
				inputEnv: map[string]string{
					"PATH": "/usr/local/bin:$PATH",
					"key":  "value",
				},
				localEnv: map[string]string{
					"A": "B",
				},
			},
			want: map[string]string{
				"PATH": "/usr/local/bin:$PATH",
				"key":  "value",
				"A":    "B",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := combineLocalAndInputEnv(tt.args.inputEnv, tt.args.localEnv); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("combineLocalAndInputEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLocalEnv(t *testing.T) {
	file1, err := ioutil.TempFile("", "file1")
	if err != nil {
		t.Fatalf("failed to to create tmpfile file1 %v", err)
	}
	defer os.Remove(file1.Name())

	content := `
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"
export SYSCONF_COMM="java"
export MAX_PROCESSORS_LIMIT="4"
export MAX_CPU_QUOTA="400"
export appDeployType="JavaWeb"
export exec_scm_hook="yes"
`
	if _, err := file1.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write content to temp file")
	}

	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "normal test case",
			args: args{filename: file1.Name()},
			want: map[string]string{
				"PATH":                 "\"/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH\"",
				"SYSCONF_COMM":         "\"java\"",
				"MAX_PROCESSORS_LIMIT": "\"4\"",
				"MAX_CPU_QUOTA":        "\"400\"",
				"appDeployType":        "\"JavaWeb\"",
				"exec_scm_hook":        "\"yes\"",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLocalEnv(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("getLocalEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLocalEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
