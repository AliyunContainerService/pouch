package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/apis/types"

	units "github.com/docker/go-units"
)

// WeightDevice defines weight device
type WeightDevice struct {
	values []*types.WeightDevice
}

func getValidateWeightDevice(val string) (*types.WeightDevice, error) {
	pairs := strings.Split(val, ":")
	if len(pairs) != 2 {
		return nil, fmt.Errorf("bad format for weight device: %s. Correct format should be <device-id>:<weight>, weight has no unit, should be in [10, 1000]", val)
	}

	weight, err := strconv.ParseUint(pairs[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid weight for device: %s", val)
	}
	if weight > 0 && (weight < 10 || weight > 1000) {
		return nil, fmt.Errorf("invalid weight for device: %s, weight should be in [10, 1000]", val)
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
	return "value"
}

func (w *WeightDevice) value() []*types.WeightDevice {
	var weightDevice []*types.WeightDevice
	for _, v := range w.values {
		weightDevice = append(weightDevice, v)
	}

	return weightDevice
}

// ThrottleBpsDevice defines throttle bps device
type ThrottleBpsDevice struct {
	values []*types.ThrottleDevice
}

func getValidThrottleDeviceBps(val string) (*types.ThrottleDevice, error) {
	pairs := strings.Split(val, ":")
	if len(pairs) != 2 {
		return nil, fmt.Errorf("bad format for throttle device: %s. Correct format should be <device-id>:<rate>, rate unit is optional and can be b/B, k/K, m/M... ", val)
	}

	rate, err := units.RAMInBytes(pairs[1])
	if err != nil || rate < 0 {
		return nil, fmt.Errorf("invalid rate for device %s, rate mustn't be negative", pairs[0])
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
	return "value"
}

func (t *ThrottleBpsDevice) value() []*types.ThrottleDevice {
	var throttleDevice []*types.ThrottleDevice
	for _, v := range t.values {
		throttleDevice = append(throttleDevice, v)
	}

	return throttleDevice
}

// ThrottleIOpsDevice defines throttle iops device
type ThrottleIOpsDevice struct {
	values []*types.ThrottleDevice
}

func getValidThrottleDeviceIOps(val string) (*types.ThrottleDevice, error) {
	pairs := strings.Split(val, ":")
	if len(pairs) != 2 {
		return nil, fmt.Errorf("bad format for throttle device: %s. Corrent format should be <device-id>:<rate>, rate unit is optional and can be b/B, k/K, m/M... ", val)
	}

	rate, err := strconv.ParseUint(pairs[1], 10, 64)
	if err != nil || rate < 0 {
		return nil, fmt.Errorf("invalid rate for device %s, rate mustn't be negative", pairs[0])
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
	return "value"
}

func (t *ThrottleIOpsDevice) value() []*types.ThrottleDevice {
	var throttleDevice []*types.ThrottleDevice
	for _, v := range t.values {
		throttleDevice = append(throttleDevice, v)
	}

	return throttleDevice
}
