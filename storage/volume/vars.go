package volume

import (
	"github.com/alibaba/pouch/storage/volume/types"
)

const (
	defaultSize       = "100G"
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
	"size":      {Value: "", Desc: "volume size"},
	"backend":   {Value: "", Desc: "volume backend"},
	"sifter":    {Value: "", Desc: "volume scheduler sifter"},
	"fs":        {Value: "ext4", Desc: "volume make filesystem"},
	"wbps":      {Value: "", Desc: "volume write bps"},
	"rbps":      {Value: "", Desc: "volume read bps"},
	"iops":      {Value: "", Desc: "volume total iops"},
	"riops":     {Value: "", Desc: "volume read iops"},
	"wiops":     {Value: "", Desc: "volume write iops"},
	"namespace": {Value: "default", Desc: "volume namespace"},
}

// ListCommonOptions returns common options.
func ListCommonOptions() map[string]types.Option {
	return commonOptions
}
