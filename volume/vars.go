package volume

import (
	"github.com/alibaba/pouch/volume/types"
)

const (
	defaultFileSystem = "ext4"
	defaultMountpoint = "/mnt"
	optionPath        = "mount"
	optionSize        = "size"
	optionSifter      = "sifter"
	optionFS          = "fs"
	optionWBps        = "wbps"
	optionRBps        = "rbps"
	optionIOps        = "iops"
	optionReadIOps    = "riops"
	optionWriteIOps   = "wiops"
	selectNamespace   = "namespace"
)

var commonOptions = map[string]types.Option{
	"size":      {"", "volume size"},
	"backend":   {"", "volume backend"},
	"sifter":    {"", "volume scheduler sifter"},
	"fs":        {"ext4", "volume make filesystem"},
	"wbps":      {"", "volume write bps"},
	"rbps":      {"", "volume read bps"},
	"iops":      {"", "volume total iops"},
	"riops":     {"", "volume read iops"},
	"wiops":     {"", "volume write iops"},
	"namespace": {"default", "volume namespace"},
}

// ListCommonOptions returns common options.
func ListCommonOptions() map[string]types.Option {
	return commonOptions
}
