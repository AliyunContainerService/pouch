package formatter

import (
	"fmt"
	"sort"
	"strings"

	"github.com/alibaba/pouch/apis/types"
	units "github.com/docker/go-units"
)

// Format key to output
const (
	TableFormat = "table"
	RawFormat   = "raw"
)

// IsTable to verify if table or raw
func IsTable(format string) bool {
	return strings.HasPrefix(format, TableFormat)
}

// PreFormat is to format the format option
func PreFormat(format string) string {
	if IsTable(format) {
		format = format[len(TableFormat):]
		// cut the space
		format = strings.Trim(format, " ")
		// input by cmd of "\t" is "\\t"
		replace := strings.NewReplacer(`\t`, "\t", `\n`, "\n")
		format = replace.Replace(format)

		if format[len(format)-1:] != "\n" {
			format += "\n"
		}
	} else {
		format = "Name:{{.Names}}\nID:{{.ID}}\nStatus:{{.Status}}\nCreated:{{.RunningFor}}\nImage:{{.Image}}\nRuntime:{{.Runtime}}\n\n"
	}
	return format
}

// LabelsToString is to transform the labels from map to string
func LabelsToString(labels map[string]string) string {
	var labelstring string
	sortedkeys := make([]string, 0)
	for key := range labels {
		sortedkeys = append(sortedkeys, key)
	}
	sort.Strings(sortedkeys)
	for _, key := range sortedkeys {
		labelstring += fmt.Sprintf("%s = %s;", key, labels[key])
	}
	return labelstring
}

// MountPointToString is to transform the MountPoint from array to string
func MountPointToString(mount []types.MountPoint) string {
	var mountstring string
	for _, value := range mount {
		mountstring += fmt.Sprintf("%s;", value.Source)
	}
	return mountstring
}

// PortBindingsToString is to transform the portbindings from map to string
func PortBindingsToString(portMap types.PortMap) string {
	var portBindings string
	sortedkeys := make([]string, 0)
	for key := range portMap {
		sortedkeys = append(sortedkeys, key)
	}
	for _, key := range sortedkeys {
		for _, ipPort := range portMap[key] {
			portBindings += fmt.Sprintf("%s->%s:%s;", key, ipPort.HostIP, ipPort.HostPort)
		}
	}
	return portBindings
}

// SizeToString is to get the size related output
func SizeToString(sizeRw int64, sizeRootFs int64) string {
	var strResult string
	sRw := units.HumanSizeWithPrecision(float64(sizeRw), 3)
	sRFs := units.HumanSizeWithPrecision(float64(sizeRootFs), 3)
	strResult = sRw
	if sizeRootFs > 0 {
		strResult = fmt.Sprintf("%s (virtual %s)", sRw, sRFs)
	}
	return strResult
}

// LocalVolumes is get the count of local volumes
func LocalVolumes(mount []types.MountPoint) string {
	c := 0
	for _, v := range mount {
		if v.Driver == "local" {
			c++
		}
	}
	return fmt.Sprintf("%d", c)
}
