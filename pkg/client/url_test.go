package client

import "testing"

func TestJoinURL(t *testing.T) {
	type args struct {
		api string
		s   []string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JoinURL(tt.args.api, tt.args.s...)
			if (err != nil) != tt.wantErr {
				t.Errorf("JoinURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("JoinURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
