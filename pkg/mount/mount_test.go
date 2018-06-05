package mount

import "testing"

func TestIsLikelyNotMountPoint(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsLikelyNotMountPoint(tt.args.file)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsLikelyNotMountPoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsLikelyNotMountPoint() = %v, want %v", got, tt.want)
			}
		})
	}
}
