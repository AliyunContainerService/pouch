package server

import "testing"

func Test_validationName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "testEmptyStr", args: args{name: ""}, wantErr: true},
		{name: "testValidationName", args: args{name: "foo"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validationName(tt.args.name); (err != nil) != tt.wantErr {
				t.Errorf("validationName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
