package config

import (
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"
)

func Test_getValidateWeightDevice(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.WeightDevice
		wantErr bool
	}{
		{
			name: "valid path and weight",
			args: args{
				val: "device1:50",
			},
			want: &types.WeightDevice{
				Path:   "device1",
				Weight: uint16(50),
			},
			wantErr: false,
		},
		{
			name: "valid path and weight",
			args: args{
				val: "device1:0",
			},
			want: &types.WeightDevice{
				Path:   "device1",
				Weight: uint16(0),
			},
			wantErr: false,
		},
		{
			name: "invalid weight -1",
			args: args{
				val: "device1:-1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid weight 5",
			args: args{
				val: "device1:5",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid upper limit 1000",
			args: args{
				val: "device1:1000",
			},
			want: &types.WeightDevice{
				Path:   "device1",
				Weight: uint16(1000),
			},
			wantErr: false,
		},
		{
			name: "invalid weight 1000 larger than upper limit",
			args: args{
				val: "device1:1001",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid weight device, and format must be <device-id>:<weight>",
			args: args{
				val: "device1",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValidateWeightDevice(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValidateWeightDevice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getValidateWeightDevice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeightDevice_Set(t *testing.T) {
	type fields struct {
		values []*types.WeightDevice
	}
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "set one valid element",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "device1:100",
			},
			wantErr: false,
		},
		{
			name: "set one invalid element with negative weight",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "device1:-100",
			},
			wantErr: true,
		},
		{
			name: "set one element already in slice",
			fields: fields{
				values: []*types.WeightDevice{
					{
						Path:   "device1",
						Weight: uint16(100),
					},
				},
			},
			args: args{
				val: "device1:100",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WeightDevice{
				values: tt.fields.values,
			}
			if err := w.Set(tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("WeightDevice.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWeightDevice_String(t *testing.T) {
	type fields struct {
		values []*types.WeightDevice
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				values: []*types.WeightDevice{
					{
						Path:   "device1",
						Weight: uint16(20),
					},
				},
			},
			want: "[device1:20]",
		},
		{
			name: "",
			fields: fields{
				values: []*types.WeightDevice{
					{
						Path:   "device1",
						Weight: uint16(20),
					},
					{
						Path:   "device2",
						Weight: uint16(40),
					},
				},
			},
			want: "[device1:20 device2:40]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WeightDevice{
				values: tt.fields.values,
			}
			if got := w.String(); got != tt.want {
				t.Errorf("WeightDevice.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeightDevice_Type(t *testing.T) {
	type fields struct {
		values []*types.WeightDevice
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "implement WeightDevice as pflag.Value interface",
			fields: fields{
				values: []*types.WeightDevice{
					{
						Path:   "device1",
						Weight: uint16(1000),
					},
				},
			},
			want: "strings",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WeightDevice{
				values: tt.fields.values,
			}
			if got := w.Type(); got != tt.want {
				t.Errorf("WeightDevice.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWeightDevice_value(t *testing.T) {
	type fields struct {
		values []*types.WeightDevice
	}
	tests := []struct {
		name   string
		fields fields
		want   []*types.WeightDevice
	}{
		{
			name: "get all values as type WeightDevice",
			fields: fields{
				values: []*types.WeightDevice{
					{
						Path:   "device1",
						Weight: uint16(1),
					},
					{
						Path:   "device2",
						Weight: uint16(1000),
					},
				},
			},
			want: []*types.WeightDevice{
				{
					Path:   "device1",
					Weight: uint16(1),
				},
				{
					Path:   "device2",
					Weight: uint16(1000),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WeightDevice{
				values: tt.fields.values,
			}
			if got := w.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WeightDevice.value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getValidThrottleDeviceBps(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ThrottleDevice
		wantErr bool
	}{
		{
			name: "valid path and weight",
			args: args{
				val: "device1:50",
			},
			want: &types.ThrottleDevice{
				Path: "device1",
				Rate: uint64(50),
			},
			wantErr: false,
		},
		{
			name: "invalid weight less than 0",
			args: args{
				val: "device1:-50",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid 50kB in rate",
			args: args{
				val: "device1:50kB",
			},
			want: &types.ThrottleDevice{
				Path: "device1",
				Rate: uint64(51200),
			},
			wantErr: false,
		},
		{
			name: "valid 500MB in rate",
			args: args{
				val: "device1:500MB",
			},
			want: &types.ThrottleDevice{
				Path: "device1",
				Rate: uint64(524288000),
			},
			wantErr: false,
		},
		{
			name: "invalid throttle device and format must be <device-id>:<rate>",
			args: args{
				val: "device1",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValidThrottleDeviceBps(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValidThrottleDeviceBps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getValidThrottleDeviceBps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThrottleBpsDevice_Set(t *testing.T) {
	type fields struct {
		values []*types.ThrottleDevice
	}
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "set one valid element",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "device1:100",
			},
			wantErr: false,
		},
		{
			name: "set one invalid element with negative weight",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "device1:-100",
			},
			wantErr: true,
		},
		{
			name: "set one element already in slice",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(100),
					},
				},
			},
			args: args{
				val: "device1:100",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttleBpsDevice := &ThrottleBpsDevice{
				values: tt.fields.values,
			}
			if err := throttleBpsDevice.Set(tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("ThrottleBpsDevice.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThrottleBpsDevice_String(t *testing.T) {
	type fields struct {
		values []*types.ThrottleDevice
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(20),
					},
				},
			},
			want: "[device1:20]",
		},
		{
			name: "",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(51200),
					},
					{
						Path: "device2",
						Rate: uint64(40),
					},
				},
			},
			want: "[device1:51200 device2:40]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttleBpsDevice := &ThrottleBpsDevice{
				values: tt.fields.values,
			}
			if got := throttleBpsDevice.String(); got != tt.want {
				t.Errorf("ThrottleBpsDevice.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThrottleBpsDevice_Type(t *testing.T) {
	type fields struct {
		values []*types.ThrottleDevice
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "implement ThrottleBpsDevice as pflag.Value interface",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(51200),
					},
				},
			},
			want: "strings",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttleBpsDevice := &ThrottleBpsDevice{
				values: tt.fields.values,
			}
			if got := throttleBpsDevice.Type(); got != tt.want {
				t.Errorf("ThrottleBpsDevice.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThrottleBpsDevice_value(t *testing.T) {
	type fields struct {
		values []*types.ThrottleDevice
	}
	tests := []struct {
		name   string
		fields fields
		want   []*types.ThrottleDevice
	}{
		{
			name: "only set Path or Rate",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
					},
					{
						Rate: uint64(51200),
					},
				},
			},
			want: []*types.ThrottleDevice{
				{
					Path: "device1",
				},
				{
					Rate: uint64(51200),
				},
			},
		},
		{
			name: "get all values as type ThrottleBpsDevice",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(51200),
					},
					{
						Path: "device2",
						Rate: uint64(102400),
					},
				},
			},
			want: []*types.ThrottleDevice{
				{
					Path: "device1",
					Rate: uint64(51200),
				},
				{
					Path: "device2",
					Rate: uint64(102400),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttleBpsDevice := &ThrottleBpsDevice{
				values: tt.fields.values,
			}
			if got := throttleBpsDevice.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ThrottleBpsDevice.value() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getValidThrottleDeviceIOps(t *testing.T) {
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ThrottleDevice
		wantErr bool
	}{
		{
			name: "valid path and weight",
			args: args{
				val: "device1:50",
			},
			want: &types.ThrottleDevice{
				Path: "device1",
				Rate: uint64(50),
			},
			wantErr: false,
		},
		{
			name: "invalid weight less than 0",
			args: args{
				val: "device1:-50",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid 50 in rate",
			args: args{
				val: "device1:50",
			},
			want: &types.ThrottleDevice{
				Path: "device1",
				Rate: uint64(50),
			},
			wantErr: false,
		},
		{
			name: "valid 500 in rate",
			args: args{
				val: "device1:500",
			},
			want: &types.ThrottleDevice{
				Path: "device1",
				Rate: uint64(500),
			},
			wantErr: false,
		},
		{
			name: "invalid throttle device and format must be <device-id>:<rate>",
			args: args{
				val: "device1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getValidThrottleDeviceIOps(tt.args.val)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValidThrottleDeviceIOps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getValidThrottleDeviceIOps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThrottleIOpsDevice_Set(t *testing.T) {
	type fields struct {
		values []*types.ThrottleDevice
	}
	type args struct {
		val string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "set one valid element",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "device1:100",
			},
			wantErr: false,
		},
		{
			name: "set one invalid element with negative weight",
			fields: fields{
				values: nil,
			},
			args: args{
				val: "device1:-100",
			},
			wantErr: true,
		},
		{
			name: "set one element already in slice",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(100),
					},
				},
			},
			args: args{
				val: "device1:100",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttleIOpsDevice := &ThrottleIOpsDevice{
				values: tt.fields.values,
			}
			if err := throttleIOpsDevice.Set(tt.args.val); (err != nil) != tt.wantErr {
				t.Errorf("ThrottleIOpsDevice.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestThrottleIOpsDevice_String(t *testing.T) {
	type fields struct {
		values []*types.ThrottleDevice
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(20),
					},
				},
			},
			want: "[device1:20]",
		},
		{
			name: "",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(51200),
					},
					{
						Path: "device2",
						Rate: uint64(40),
					},
				},
			},
			want: "[device1:51200 device2:40]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttleIOpsDevice := &ThrottleIOpsDevice{
				values: tt.fields.values,
			}
			if got := throttleIOpsDevice.String(); got != tt.want {
				t.Errorf("ThrottleIOpsDevice.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThrottleIOpsDevice_Type(t *testing.T) {
	type fields struct {
		values []*types.ThrottleDevice
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "implement ThrottleIOpsDevice as pflag.Value interface",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(51200),
					},
				},
			},
			want: "strings",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttleIOpsDevice := &ThrottleIOpsDevice{
				values: tt.fields.values,
			}
			if got := throttleIOpsDevice.Type(); got != tt.want {
				t.Errorf("ThrottleIOpsDevice.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThrottleIOpsDevice_value(t *testing.T) {
	type fields struct {
		values []*types.ThrottleDevice
	}
	tests := []struct {
		name   string
		fields fields
		want   []*types.ThrottleDevice
	}{
		{
			name: "only set Path or Rate",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
					},
					{
						Rate: uint64(102400),
					},
				},
			},
			want: []*types.ThrottleDevice{
				{
					Path: "device1",
				},
				{
					Rate: uint64(102400),
				},
			},
		},
		{
			name: "get all values as type ThrottleIOpsDevice",
			fields: fields{
				values: []*types.ThrottleDevice{
					{
						Path: "device1",
						Rate: uint64(51200),
					},
					{
						Path: "device2",
						Rate: uint64(102400),
					},
				},
			},
			want: []*types.ThrottleDevice{
				{
					Path: "device1",
					Rate: uint64(51200),
				},
				{
					Path: "device2",
					Rate: uint64(102400),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			throttleIOpsDevice := &ThrottleIOpsDevice{
				values: tt.fields.values,
			}
			if got := throttleIOpsDevice.Value(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ThrottleIOpsDevice.value() = %v, want %v", got, tt.want)
			}
		})
	}
}
