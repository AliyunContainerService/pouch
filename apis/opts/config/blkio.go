package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	units "github.com/docker/go-units"
)

const blkioOptsType = "strings"

// WeightDevice defines weight device
type WeightDevice struct {
	values []*types.WeightDevice
}

func getValidateWeightDevice(val string) (*types.WeightDevice, error) {
	pairs := strings.Split(val, ":")
	if len(pairs) != 2 {
		return nil, fmt.Errorf("invalid weight device %s: format must be <device-id>:<weight> with weight in range [10, 1000]", val)
	}

	weight, err := strconv.ParseUint(pairs[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid weight device %s: weight cannot be less than 0", val)
	}
	if weight > 0 && (weight < 10 || weight > 1000) {
		return nil, fmt.Errorf("invalid weight device %s: weight must be in range [10, 1000]", val)
	}

	return &types.WeightDevice{
		Path:   pairs[0],
		Weight: uint16(weight),
	}, nil
}

// Set implement WeightDevice as pflag.Value interface
func (w *WeightDevice) Set(val string) error {
	v, err := getValidateWeightDevice(val)
	if err != nil {
		return err
	}
	w.values = append(w.values, v)

	return nil
}

// String implement WeightDevice as pflag.Value interface
func (w *WeightDevice) String() string {
	var str []string
	for _, v := range w.values {
		str = append(str, fmt.Sprintf("%s:%d", v.Path, v.Weight))
	}

	return fmt.Sprintf("%v", str)
}

// Type implement WeightDevice as pflag.Value interface
func (w *WeightDevice) Type() string {
	return blkioOptsType
}

// Value returns all values as type WeightDevice
func (w *WeightDevice) Value() []*types.WeightDevice {
	var weightDevice []*types.WeightDevice
	weightDevice = append(weightDevice, w.values...)

	return weightDevice
}

// ThrottleBpsDevice defines throttle bps device
type ThrottleBpsDevice struct {
	values []*types.ThrottleDevice
}

func getValidThrottleDeviceBps(val string) (*types.ThrottleDevice, error) {
	pairs := strings.Split(val, ":")
	if len(pairs) != 2 {
		return nil, fmt.Errorf("invalid throttle device %s: format must be <device-id>:<rate> with optional rate unit", val)
	}

	rate, err := units.RAMInBytes(pairs[1])
	if err != nil || rate < 0 {
		return nil, fmt.Errorf("invalid rate %s for device %s: cannot be negative", pairs[1], pairs[0])
	}

	return &types.ThrottleDevice{
		Path: pairs[0],
		Rate: uint64(rate),
	}, nil
}

// Set implement ThrottleBpsDevice as pflag.Value interface
func (t *ThrottleBpsDevice) Set(val string) error {
	v, err := getValidThrottleDeviceBps(val)
	if err != nil {
		return err
	}
	t.values = append(t.values, v)

	return nil
}

// String implement ThrottleBpsDevice as pflag.Value interface
func (t *ThrottleBpsDevice) String() string {
	var str []string
	for _, v := range t.values {
		str = append(str, fmt.Sprintf("%s:%d", v.Path, v.Rate))
	}

	return fmt.Sprintf("%v", str)
}

// Type implement ThrottleBpsDevice as pflag.Value interface
func (t *ThrottleBpsDevice) Type() string {
	return blkioOptsType
}

// Value returns all values as type ThrottleDevice
func (t *ThrottleBpsDevice) Value() []*types.ThrottleDevice {
	var throttleDevice []*types.ThrottleDevice
	throttleDevice = append(throttleDevice, t.values...)

	return throttleDevice
}

// ThrottleIOpsDevice defines throttle iops device
type ThrottleIOpsDevice struct {
	values []*types.ThrottleDevice
}

func getValidThrottleDeviceIOps(val string) (*types.ThrottleDevice, error) {
	pairs := strings.Split(val, ":")
	if len(pairs) != 2 {
		return nil, fmt.Errorf("invalid throttle device %s: format must be <device-id>:<rate> with optional rate unit", val)
	}

	rate, err := strconv.ParseUint(pairs[1], 10, 64)
	if err != nil || rate < 0 {
		return nil, fmt.Errorf("invalid rate %s for device %s: rate cannot be negative", pairs[1], pairs[0])
	}

	return &types.ThrottleDevice{
		Path: pairs[0],
		Rate: uint64(rate),
	}, nil
}

// Set implement ThrottleIOpsDevice as pflag.Value interface
func (t *ThrottleIOpsDevice) Set(val string) error {
	v, err := getValidThrottleDeviceIOps(val)
	if err != nil {
		return err
	}
	t.values = append(t.values, v)

	return nil
}

// String implement ThrottleIOpsDevice as pflag.Value interface
func (t *ThrottleIOpsDevice) String() string {
	var str []string
	for _, v := range t.values {
		str = append(str, fmt.Sprintf("%s:%d", v.Path, v.Rate))
	}

	return fmt.Sprintf("%v", str)
}

// Type implement ThrottleIOpsDevice as pflag.Value interface
func (t *ThrottleIOpsDevice) Type() string {
	return blkioOptsType
}

// Value returns all values
func (t *ThrottleIOpsDevice) Value() []*types.ThrottleDevice {
	var throttleDevice []*types.ThrottleDevice
	throttleDevice = append(throttleDevice, t.values...)

	return throttleDevice
}
