package types

const (
	defaultSize       = "100G"
	defaultFileSystem = "ext4"
	defaultMountpoint = "/mnt"
	optionPath        = "mount"
	optionSize        = "opt.size"
	optionSifter      = "sifter"
	optionFS          = "opt.fs"
	optionWBps        = "opt.wbps"
	optionRBps        = "opt.rbps"
	optionIOps        = "opt.iops"
	optionReadIOps    = "opt.riops"
	optionWriteIOps   = "opt.wiops"
	selectNamespace   = "namespace"
)

var commonOptions = map[string]Option{
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
func ListCommonOptions() map[string]Option {
	return commonOptions
}
