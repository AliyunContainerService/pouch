package opts

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"
)

func TestParseNetworks(t *testing.T) {
	type args struct {
		networks []string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.NetworkingConfig
		want1   string
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ParseNetworks(tt.args.networks)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNetworks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseNetworks() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseNetworks() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestVerifyNetworks(t *testing.T) {
	type args struct {
		nwConfig *types.NetworkingConfig
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
			if err := ValidateNetworks(tt.args.nwConfig); (err != nil) != tt.wantErr {
				t.Errorf("VerifyNetworks() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parseNetwork(t *testing.T) {
	type net struct {
		name      string
		parameter string
		mode      string
	}
	type result struct {
		network net
		err     error
	}
	type TestCases struct {
		input  string
		expect result
	}

	testCases := []TestCases{
		{
			input: "",
			expect: result{
				err:     fmt.Errorf("invalid network: cannot be empty"),
				network: net{name: "", parameter: "", mode: ""},
			},
		},
		{
			input: "121.0.0.1",
			expect: result{
				err:     nil,
				network: net{name: "", parameter: "121.0.0.1", mode: ""},
			},
		},
		{
			input: "myHost",
			expect: result{
				err:     nil,
				network: net{name: "myHost", parameter: "", mode: ""},
			},
		},
		{
			input: "myHost:121.0.0.1",
			expect: result{
				err:     nil,
				network: net{name: "myHost", parameter: "121.0.0.1", mode: ""},
			},
		},
		{
			input: "container:9ca6ac",
			expect: result{
				err:     nil,
				network: net{name: "container", parameter: "9ca6ac", mode: ""},
			},
		},
		{
			input: "bridge:121.0.0.1:mode",
			expect: result{
				err:     nil,
				network: net{name: "bridge", parameter: "121.0.0.1", mode: "mode"},
			},
		},
		{
			input: "bridge:mode",
			expect: result{
				err:     nil,
				network: net{name: "bridge", parameter: "", mode: "mode"},
			},
		},
	}

	for _, testCase := range testCases {
		name, parameter, mode, error := parseNetwork(testCase.input)
		assert.Equal(t, testCase.expect.err, error)
		assert.Equal(t, testCase.expect.network.name, name)
		assert.Equal(t, testCase.expect.network.parameter, parameter)
		assert.Equal(t, testCase.expect.network.mode, mode)
	}
}
