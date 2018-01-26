package exec

import (
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	type args struct {
		timeout time.Duration
		bin     string
		args    []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   string
		want2   string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := Run(tt.args.timeout, tt.args.bin, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Run() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Run() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("Run() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestRunWithRetry(t *testing.T) {
	type args struct {
		times    int
		interval time.Duration
		timeout  time.Duration
		bin      string
		args     []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   string
		want2   string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, err := RunWithRetry(tt.args.times, tt.args.interval, tt.args.timeout, tt.args.bin, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunWithRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RunWithRetry() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("RunWithRetry() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("RunWithRetry() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestRetry(t *testing.T) {
	type args struct {
		times    int
		interval time.Duration
		f        func() error
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
			if err := Retry(tt.args.times, tt.args.interval, tt.args.f); (err != nil) != tt.wantErr {
				t.Errorf("Retry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
