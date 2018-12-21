package opts

import "testing"

func TestParseIntelRdt(t *testing.T) {
	type args struct {
		intelRdtL3Cbm string
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
			got, err := ParseIntelRdt(tt.args.intelRdtL3Cbm)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIntelRdt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseIntelRdt() = %v, want %v", got, tt.want)
			}
		})
	}
}
