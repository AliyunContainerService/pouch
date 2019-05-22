package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"
	criconfig "github.com/alibaba/pouch/cri/config"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/storage/volume"

	"github.com/containerd/containerd/namespaces"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestFlattenConfig(t *testing.T) {
	assert := assert.New(t)
	origin := map[string]interface{}{
		"a": "a",
		"b": "b",
		"c": "c",
		"iter1": map[string]interface{}{
			"i1": "i1",
			"i2": "i2",
		},
		"iter11": map[string]interface{}{
			"ii1": map[string]interface{}{
				"iii1": "iii1",
				"iii2": "iii2",
			},
		},
	}

	expect := map[string]interface{}{
		"a":  "a",
		"b":  "b",
		"c":  "c",
		"i1": "i1",
		"i2": "i2",
		"ii1": map[string]interface{}{
			"iii1": "iii1",
			"iii2": "iii2",
		},
	}

	config := make(map[string]interface{})
	flattenConfig(origin, config)
	assert.Equal(config, expect)

	// test nil map will not cause panic
	config = make(map[string]interface{})
	flattenConfig(nil, config)
	assert.Equal(config, map[string]interface{}{})
}

func TestConfigValidate(t *testing.T) {
	assert := assert.New(t)
	// Test default configuration
	cfg := &Config{}
	assert.Equal(nil, cfg.Validate())

	// Test volume configuration
	cfg = &Config{
		VolumeConfig: volume.Config{
			RemoveVolume:   false,
			DefaultBackend: "/path/to",
			VolumeMetaPath: "/foo/bar",
			DriverAlias:    "local=alilocal",
		},
	}
	assert.Equal(nil, cfg.Validate())

	// Test network configuration
	cfg = &Config{
		NetworkConfig: network.Config{
			MetaPath:   "/path/to",
			ExecRoot:   "exec-root-dir",
			DNS:        []string{},
			DNSOptions: []string{},
			BridgeConfig: network.BridgeConfig{
				Name:          "net-config-test",
				GatewayIPv4:   "192.168.5.1",
				PreferredIP:   "",
				Mtu:           1500,
				ICC:           false,
				IPTables:      true,
				IPForward:     true,
				IPMasq:        false,
				UserlandProxy: false,
			},
		},
	}
	assert.Equal(nil, cfg.Validate())

	// Test cri configuration
	cfg = &Config{
		IsCriEnabled: true,
		CriConfig: criconfig.Config{
			Listen:                "unix:///var/run/pouchd.sock",
			NetworkPluginBinDir:   "cni-bin-dir",
			NetworkPluginConfDir:  "cni-conf-dir",
			SandboxImage:          "registry.hub.docker.com/library/busybox",
			CriVersion:            "v1alpha2",
			StreamServerPort:      "10010",
			StreamServerReusePort: true,
			CriStatsCollectPeriod: 10,
			EnableCriStatsCollect: false,
		},
	}
	assert.Equal(nil, cfg.Validate())

	// Test registry configuration
	cfg = &Config{
		DefaultRegistry:   "registry.hub.docker.com",
		DefaultRegistryNS: "library",
	}
	assert.Equal(nil, cfg.Validate())

	// Test TLS configuration
	cfg = &Config{
		TLS: client.TLSConfig{
			CA:               "/path/to/.pouchcert/ca.pem",
			Cert:             "/path/to/.pouchcert/cert.pem",
			Key:              "/path/to/.pouchcert/key.pem",
			VerifyRemote:     true,
			ManagerWhiteList: "docker.alibaba.com",
		},
	}
	assert.Equal(nil, cfg.Validate())

	// Test lxcfs configuration
	cfg = &Config{
		IsLxcfsEnabled: false,
		LxcfsBinPath:   "/usr/local/bin/lxcfs",
		LxcfsHome:      "/var/lib/lxcfs",
	}
	assert.Equal(nil, cfg.Validate())

	// Test label configuration
	cfg = &Config{
		Labels: []string{},
	}
	assert.Equal(nil, cfg.Validate())

	cfg = &Config{
		Labels: []string{
			"a=b",
		},
	}
	assert.Equal(nil, cfg.Validate())

	cfg = &Config{
		Labels: []string{
			"a=b",
			"c=d",
		},
	}
	assert.Equal(nil, cfg.Validate())

	cfg = &Config{
		Labels: []string{
			"foo=bar",
		},
	}
	assert.Equal(nil, cfg.Validate())

	// Test others configuration
	cfg = &Config{
		Debug: true,
		Listen: []string{
			"unix:///var/run/pouchd.sock",
		},
		ContainerdAddr:   "/var/run/containerd.sock",
		HomeDir:          "/var/lib/pouch",
		ContainerdPath:   "/usr/local/bin",
		DefaultRuntime:   "runc",
		ImageProxy:       "",
		QuotaDriver:      "grpquota",
		ConfigFile:       "/etc/pouch/config.json",
		CgroupParent:     "default",
		EnableProfiler:   true,
		Pidfile:          "/var/run/pouch.pid",
		OOMScoreAdjust:   -500,
		DefaultNamespace: namespaces.Default,
		DefaultLogConfig: types.LogConfig{
			LogDriver: types.LogConfigLogDriverJSONFile,
			LogOpts:   map[string]string{},
		},
	}
	assert.Equal(nil, cfg.Validate())
}

func TestGetConflictConfigurations(t *testing.T) {
	assert := assert.New(t)

	fileflags := map[string]interface{}{
		"a": "a1",
		"b": []string{"b1", "b2"},
	}

	flags := pflag.NewFlagSet("cmflags", pflag.ContinueOnError)

	// Test No Flags
	assert.Equal(nil, getConflictConfigurations(flags, fileflags))

	flags.String("c", "c1", "c")
	// Test No Conflicts
	flags.Parse([]string{"--c=c1"})
	assert.Equal(nil, getConflictConfigurations(flags, fileflags))

	// Test Ignore Conflict of Type "Slice"
	flags.StringSlice("b", []string{"b1", "b2"}, "b")
	flags.Parse([]string{"--b=b1,b2"})
	assert.Equal(nil, getConflictConfigurations(flags, fileflags))

	// Test Conflict
	flags.String("a", "a1", "a")
	flags.Parse([]string{"--a=a1"})
	assert.Equal("found conflict flags in command line and config file: from flag: a1 and from config file: a1",
		getConflictConfigurations(flags, fileflags).Error())
}

func TestMergeConfigurations(t *testing.T) {
	assert := assert.New(t)
	flags := pflag.NewFlagSet("cmflags", pflag.ContinueOnError)

	// Test configuration file doesn't exist
	cfg := &Config{
		ConfigFile: "/tmp/non-existent.json",
	}

	assert.Equal(nil, cfg.MergeConfigurations(flags))

	// Test invalid content format
	configFile := "/tmp/test-merge-config-1.json"
	err := ioutil.WriteFile(configFile, []byte(""), 0644)
	assert.Equal(nil, err)

	cfg = &Config{
		ConfigFile: configFile,
	}

	err = cfg.MergeConfigurations(flags)
	if err == nil || !strings.Contains(err.Error(), "failed to decode json") {
		t.Errorf("error string '%v' is not wantted", err)
	}

	defer os.Remove(configFile)

	// Test correct configuration
	configFile = "/tmp/test-merge-config-2.json"
	config := struct {
		Labels []string `json:"label,omitempty"`
	}{
		Labels: []string{"foo=bar"},
	}

	body, errs := json.Marshal(config)
	assert.NoError(errs)

	err = ioutil.WriteFile(configFile, body, 0644)
	assert.NoError(err)

	cfg = &Config{
		ConfigFile: configFile,
	}

	err = cfg.MergeConfigurations(flags)
	assert.NoError(err)

	defer os.Remove(configFile)
}

func TestValidateCgroupDriver(t *testing.T) {
	for _, tc := range []struct {
		driver    string
		expectErr bool
	}{
		{
			driver:    CgroupfsDriver,
			expectErr: false,
		},
		{
			driver:    CgroupSystemdDriver,
			expectErr: false,
		},
		{
			driver:    "foo",
			expectErr: true,
		},
	} {
		err := validateCgroupDriver(tc.driver)
		if tc.expectErr != (err != nil) {
			t.Fatalf("expectd error: %v, but get %s", tc.expectErr, err)
		}
	}
}
