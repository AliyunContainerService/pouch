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
		err     error
		wantErr bool
	}{
		{
			name: "networks is null then return 'bridge' mode",
			args: args{
				networks: []string{},
			},
			want: &types.NetworkingConfig{
				EndpointsConfig: map[string]*types.EndpointSettings{},
			},
			want1:   "bridge",
			wantErr: false,
		},
		{
			name: "invalid network: cannot be empty",
			args: args{
				networks: []string{""},
			},
			err:     fmt.Errorf("invalid network: cannot be empty"),
			wantErr: true,
		},
		{
			name: "if networkMode is null or mode is 'mode' then return name as mode",
			args: args{
				networks: []string{"foo:bar:mode"},
			},
			want: &types.NetworkingConfig{
				EndpointsConfig: map[string]*types.EndpointSettings{},
			},
			want1:   "foo",
			wantErr: false,
		},
		{
			name: "if name is 'container' then return name and parameter as mode",
			args: args{
				networks: []string{"container:e8e153651a0d:mode"},
			},
			want: &types.NetworkingConfig{
				EndpointsConfig: map[string]*types.EndpointSettings{},
			},
			want1:   "container:e8e153651a0d",
			wantErr: false,
		},
		{
			name: "name is not 'container'",
			args: args{
				networks: []string{"foo:127.0.0.1"},
			},
			want: &types.NetworkingConfig{
				EndpointsConfig: map[string]*types.EndpointSettings{
					"foo": {
						IPAddress: "127.0.0.1",
						IPAMConfig: &types.EndpointIPAMConfig{
							IPV4Address: "127.0.0.1",
						},
					},
				},
			},
			want1:   "foo",
			wantErr: false,
		},
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
			if (err != nil) && !reflect.DeepEqual(err, tt.err) {
				t.Errorf("ParseNetworks() = %v, want %v", err, tt.err)
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
		err     error
		wantErr bool
	}{
		{
			name: "return expected nil 1",
			args: args{
				nwConfig: &types.NetworkingConfig{},
			},
			wantErr: false,
		},
		{
			name: "return expected nil 2",
			args: args{
				nwConfig: &types.NetworkingConfig{
					EndpointsConfig: map[string]*types.EndpointSettings{},
				},
			},
			wantErr: false,
		},
		{
			name: "test IPAMConfig.IPV4Address",
			args: args{
				nwConfig: &types.NetworkingConfig{
					EndpointsConfig: map[string]*types.EndpointSettings{
						"Networks": {
							IPAMConfig: &types.EndpointIPAMConfig{
								IPV4Address: "127.0.0.1",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "test IPAddress, Gateway, IPV4Address and IPV6Address",
			args: args{
				nwConfig: &types.NetworkingConfig{
					EndpointsConfig: map[string]*types.EndpointSettings{
						"Networks": {
							Links:      []string{},
							Aliases:    []string{},
							EndpointID: "",
							NetworkID:  "",
							MacAddress: "",
							IPAddress:  "127.0.0.1",
							Gateway:    "192.168.1.1",
							IPAMConfig: &types.EndpointIPAMConfig{
								IPV4Address: "192.168.1.1",
								IPV6Address: "2001:0db8:85a3:08d3:1319:8a2e:0370:1234",
								LinkLocalIps: []string{
									"1.2.3.4",
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "container cannot be connected to network endpoints",
			args: args{
				nwConfig: &types.NetworkingConfig{
					EndpointsConfig: map[string]*types.EndpointSettings{
						"foo": {},
						"bar": {},
					},
				},
			},
			wantErr: true,
			err:     fmt.Errorf("Container cannot be connected to network endpoints: bar, foo"),
		},
		{
			name: "invalid IPv4 address",
			args: args{
				nwConfig: &types.NetworkingConfig{
					EndpointsConfig: map[string]*types.EndpointSettings{
						"foo": {
							IPAMConfig: &types.EndpointIPAMConfig{
								IPV4Address: "foo.bar",
							},
						},
					},
				},
			},
			wantErr: true,
			err:     fmt.Errorf("invalid IPv4 address: foo.bar"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNetworks(tt.args.nwConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyNetworks() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !reflect.DeepEqual(err, tt.err) {
				t.Errorf("VerifyNetworks() = %v, want %v", err, tt.err)
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
