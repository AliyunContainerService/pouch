package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/client"
	criconfig "github.com/alibaba/pouch/cri/config"
	"github.com/alibaba/pouch/network"
	"github.com/alibaba/pouch/pkg/utils"
	"github.com/alibaba/pouch/storage/volume"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

const (
	// CgroupfsDriver is cgroupfs driver
	CgroupfsDriver = "cgroupfs"
	// CgroupSystemdDriver is systemd driver
	CgroupSystemdDriver = "systemd"
	// DefaultCgroupDriver is default cgroups driver
	DefaultCgroupDriver = CgroupfsDriver
)

// Config refers to daemon's whole configurations.
type Config struct {
	sync.Mutex `json:"-"`

	//Volume config
	VolumeConfig volume.Config `json:"volume-config,omitempty"`

	// Network config
	NetworkConfig network.Config `json:"network-config,omitempty"`

	// Whether enable cri manager.
	IsCriEnabled bool `json:"enable-cri,omitempty"`

	// CRI config.
	CriConfig criconfig.Config `json:"cri-config,omitempty"`

	// Server listening address.
	Listen []string `json:"listen,omitempty"`

	// Debug refers to the log mode.
	Debug bool `json:"debug,omitempty"`

	// ContainerdAddr refers to the unix socket path of containerd.
	ContainerdAddr string `json:"containerd,omitempty"`

	// DefaultRegistry is daemon's default registry which is to pull/push/search images.
	DefaultRegistry string `json:"default-registry,omitempty"`

	// DefaultRegistryNS is daemon's default registry namespace used in pull/push/search images.
	DefaultRegistryNS string `json:"default-registry-namespace,omitempty"`

	// Home directory.
	HomeDir string `json:"home-dir,omitempty"`

	// ContainerdPath is the absolute path of containerd binary,
	// /usr/local/bin is the default.
	ContainerdPath string `json:"containerd-path,omitempty"`

	// TLS configuration
	TLS client.TLSConfig `json:"TLS,omitempty"`

	// Default OCI Runtime
	DefaultRuntime string `json:"default-runtime,omitempty"`

	// Enable lxcfs
	IsLxcfsEnabled bool `json:"enable-lxcfs,omitempty"`

	// LxcfsBinPath is the absolute path of lxcfs binary
	LxcfsBinPath string `json:"lxcfs,omitempty"`

	// LxcfsHome is the absolute path of lxcfs
	LxcfsHome string `json:"lxcfs-home,omitempty"`

	// ImagxeProxy is a http proxy to pull image
	ImageProxy string `json:"image-proxy,omitempty"`

	// QuotaDriver is used to set the driver of Quota
	QuotaDriver string `json:"quota-driver,omitempty"`

	// Configuration file of pouchd
	ConfigFile string `json:"config-file,omitempty"`

	// CgroupParent is to set parent cgroup for all containers
	CgroupParent string `json:"cgroup-parent,omitempty"`

	// Labels is the metadata of daemon
	Labels []string `json:"label,omitempty"`

	// EnableProfiler indicates whether pouchd setup profiler like pprof and stack dumping etc
	EnableProfiler bool `json:"enable-profiler,omitempty"`

	// Pidfile keeps daemon pid
	Pidfile string `json:"pidfile,omitempty"`

	// Default log configuration
	DefaultLogConfig types.LogConfig `json:"default-log-config,omitempty"`

	// RegistryService
	RegistryService types.RegistryServiceConfig `json:"registry-service,omitempty" `

	// oom_score_adj for the daemon
	OOMScoreAdjust int `json:"oom-score-adjust,omitempty"`

	// runtimes config
	Runtimes map[string]types.Runtime `json:"add-runtime,omitempty"`

	// DefaultNamespace is passed to containerd.
	DefaultNamespace string `json:"default-namespace,omitempty"`

	// Snapshotter is passed to containerd, default to overlayfs
	Snapshotter string `json:"snapshotter,omitempty"`

	// AllowMultiSnapshotter allows multi snapshotter, default false
	AllowMultiSnapshotter bool `json:"allow-multi-snapshotter,omitempty"`

	// CgroupDriver sets cgroup driver for all containers
	CgroupDriver string `json:"cgroup-driver,omitempty"`

	// InsecureRegistries sets insecure registries to allow to pull
	// insecure registries.
	InsecureRegistries []string `json:"insecure-registries,omitempty"`
}

// GetCgroupDriver gets cgroup driver used in runc.
func (cfg *Config) GetCgroupDriver() string {
	return cfg.CgroupDriver
}

// UseSystemd tells whether use systemd cgroup driver
func (cfg *Config) UseSystemd() bool {
	return cfg.CgroupDriver == CgroupSystemdDriver
}

// Validate validates the user input config.
func (cfg *Config) Validate() error {
	// for debug config file.
	b, _ := json.Marshal(cfg)
	logrus.Debugf("daemon config: (%s)", string(b))

	// deduplicated elements in slice if there is any.
	cfg.Listen = utils.DeDuplicate(cfg.Listen)
	cfg.Labels = utils.DeDuplicate(cfg.Labels)

	for _, label := range cfg.Labels {
		data := strings.SplitN(label, "=", 2)
		if len(data) != 2 {
			return fmt.Errorf("daemon label %s must be in format of key=value", label)
		}
		if len(data[0]) == 0 || len(data[1]) == 0 {
			return fmt.Errorf("key and value in daemon label %s cannot be empty", label)
		}
	}

	// TODO: add config validation

	// validates runtimes config
	if len(cfg.Runtimes) == 0 {
		cfg.Runtimes = make(map[string]types.Runtime)
	}
	if _, exist := cfg.Runtimes[cfg.DefaultRuntime]; !exist {
		// add default runtime
		cfg.Runtimes[cfg.DefaultRuntime] = types.Runtime{Path: cfg.DefaultRuntime}
	}

	// if cgroup driver is empty, use default cgroup driver
	if cfg.CgroupDriver == "" {
		cfg.CgroupDriver = DefaultCgroupDriver
	}

	return validateCgroupDriver(cfg.CgroupDriver)
}

//MergeConfigurations merges flagSet flags and config file flags into Config.
func (cfg *Config) MergeConfigurations(flagSet *pflag.FlagSet) error {
	contents, err := ioutil.ReadFile(cfg.ConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			logrus.Debugf("the %v doesn't exist: %v", cfg.ConfigFile, err)
			return nil
		}
		return fmt.Errorf("failed to read contents from config file %s: %s", cfg.ConfigFile, err)
	}

	var origin map[string]interface{}
	if err = json.NewDecoder(bytes.NewReader(contents)).Decode(&origin); err != nil {
		return fmt.Errorf("failed to decode json: %s", err)
	}
	if len(origin) == 0 {
		return nil
	}

	fileFlags := make(map[string]interface{})
	flattenConfig(origin, fileFlags)

	// check conflict in command line flags and config file
	if err = getConflictConfigurations(flagSet, fileFlags); err != nil {
		return err
	}

	fileConfig := &Config{}
	if err = json.NewDecoder(bytes.NewReader(contents)).Decode(fileConfig); err != nil {
		return fmt.Errorf("failed to decode json: %s", err)
	}

	// merge configurations from command line flags and config file
	return mergeConfigurations(fileConfig, cfg.delValue(flagSet, fileFlags))
}

// delValue deleles value in config, since we do not do conflict check for slice
// type flag, note that we should remove default flag value in merging, cause
// this is not reasonable if the flag is not passed. Just set the flag value to
// null, when same flag has been set in config file.
func (cfg *Config) delValue(flagSet *pflag.FlagSet, fileFlags map[string]interface{}) *Config {
	flagSet.VisitAll(func(f *pflag.Flag) {
		// if flag type not slice or array , then skip
		if !strings.Contains(f.Value.Type(), "Slice") && !strings.Contains(f.Value.Type(), "Array") {
			return
		}

		// if flag is set in command line, then skip
		if f.Changed {
			return
		}

		// if flag is not set in config file, then skip
		if _, exist := fileFlags[f.Name]; !exist {
			return
		}

		// set value as null in config
		r := reflect.ValueOf(cfg).Elem()
		rtype := r.Type()
		for i := 0; i < r.NumField(); i++ {
			if rtype.Field(i).Type.Kind() != reflect.Slice {
				continue
			}
			if strings.Contains(rtype.Field(i).Tag.Get("json"), f.Name) {
				r.Field(i).Set(reflect.MakeSlice(reflect.TypeOf([]string{}), 0, 0))
			}
		}
	})

	return cfg
}

// flattenConfig flattens key-value config.
func flattenConfig(origin map[string]interface{}, config map[string]interface{}) {
	for k, v := range origin {
		if c, ok := v.(map[string]interface{}); ok {
			for fk, fv := range c {
				config[fk] = fv
			}
		} else {
			config[k] = v
		}
	}
}

// find conflict in command line flags and config file, note that if flag value
// is slice type, we will skip it and merge it from flags and config file later.
func getConflictConfigurations(flagSet *pflag.FlagSet, fileFlags map[string]interface{}) error {
	var conflictFlags []string
	flagSet.Visit(func(f *pflag.Flag) {
		flagType := f.Value.Type()
		if strings.Contains(flagType, "Slice") {
			return
		}
		if v, exist := fileFlags[f.Name]; exist {
			conflictFlags = append(conflictFlags, fmt.Sprintf("from flag: %s and from config file: %s", f.Value.String(), v))
		}
	})

	if len(conflictFlags) > 0 {
		return fmt.Errorf("found conflict flags in command line and config file: %v", strings.Join(conflictFlags, ", "))
	}

	return nil
}

// merge flagSet and config file into cfg
func mergeConfigurations(src *Config, dest *Config) error {
	return utils.Merge(src, dest)
}

// validateCgroupDriver validates cgroup driver
func validateCgroupDriver(driver string) error {
	if driver == CgroupfsDriver || driver == CgroupSystemdDriver {
		return nil
	}

	return fmt.Errorf("invalid cgroup driver: %s, valid driver is cgroupfs or systemd", driver)
}
