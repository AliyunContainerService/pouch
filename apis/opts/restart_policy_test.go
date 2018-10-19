package opts

import (
	"fmt"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"
)

func TestParseRestartPolicy(t *testing.T) {
	type TestCase struct {
		input         string
		expectedName  string
		expectedCount int64
		err           error
	}

	cases := []TestCase{
		{
			input:         "always",
			expectedName:  "always",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "no",
			expectedName:  "no",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "unless-stopped",
			expectedName:  "unless-stopped",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "on-failure:1",
			expectedName:  "on-failure",
			expectedCount: 1,
			err:           nil,
		},
		{
			input:         "on-failure",
			expectedName:  "on-failure",
			expectedCount: 0,
			err:           nil,
		},
		{
			input:         "on-failure:1:2",
			expectedName:  "on-failure",
			expectedCount: 0,
			err:           fmt.Errorf("invalid restart policy: %s", "on-failure:1:2"),
		},
		{
			input:         "",
			expectedName:  "no",
			expectedCount: 0,
		},
		{
			input:         "on-failure:foo",
			expectedName:  "on-failure",
			expectedCount: 0,
			err:           fmt.Errorf("invalid restart policy: strconv.Atoi: parsing \"foo\": invalid syntax"),
		},
		{
			input:         "default",
			expectedName:  "nil",
			expectedCount: 0,
			err:           fmt.Errorf("invalid restart policy: default"),
		},
	}

	for _, cs := range cases {
		policy, err := ParseRestartPolicy(cs.input)
		assert.Equal(t, cs.err, err)
		if err == nil {
			assert.Equal(t, cs.expectedName, policy.Name)
			assert.Equal(t, cs.expectedCount, policy.MaximumRetryCount)
		}
	}
}

func TestValidateRestartPolicy(t *testing.T) {
	type args struct {
		policy *types.RestartPolicy
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{name: "test1", args: args{policy: &types.RestartPolicy{Name: "always", MaximumRetryCount: 2}}, wantErr: true},
		{name: "test2", args: args{policy: &types.RestartPolicy{Name: "unless-stopped", MaximumRetryCount: 2}}, wantErr: true},
		{name: "test3", args: args{policy: &types.RestartPolicy{Name: "no", MaximumRetryCount: 2}}, wantErr: true},
		{name: "test4", args: args{policy: &types.RestartPolicy{Name: "on-failure", MaximumRetryCount: -1}}, wantErr: true},
		{name: "test5", args: args{policy: &types.RestartPolicy{Name: "", MaximumRetryCount: 2}}, wantErr: false},
		{name: "test6", args: args{policy: &types.RestartPolicy{Name: "foo", MaximumRetryCount: 2}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateRestartPolicy(tt.args.policy); (err != nil) != tt.wantErr {
				t.Errorf("VerifyRestartPolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
