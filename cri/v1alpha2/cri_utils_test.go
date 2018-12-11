package v1alpha2

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	apitypes "github.com/alibaba/pouch/apis/types"
	runtime "github.com/alibaba/pouch/cri/apis/v1alpha2"
	"github.com/alibaba/pouch/daemon/mgr"
	"github.com/alibaba/pouch/pkg/utils"

	"github.com/cri-o/ocicni/pkg/ocicni"
	"github.com/stretchr/testify/assert"
)

var (
	memorySwappiness           int64 = 1000
	apitypesWeightDevicesSlice       = []*apitypes.WeightDevice{
		{
			Path:   "foo",
			Weight: uint16(1),
		},
		{
			Path:   "foo2",
			Weight: uint16(2),
		},
	}
	apitypesThrottleDevicesSlice = []*apitypes.ThrottleDevice{
		{
			Path: "foo",
			Rate: uint64(1000),
		},
		{
			Path: "foo2",
			Rate: uint64(2000),
		},
	}
	apitypesUlimitsSlice = []*apitypes.Ulimit{
		{
			Name: "foo",
			Hard: 1,
			Soft: 1,
		},
		{
			Name: "foo2",
			Hard: 2,
			Soft: 2,
		},
		{
			Name: "foo3",
			Hard: -1,
			Soft: -1,
		},
	}
	runtimeWeightDevicesSlice = []*runtime.WeightDevice{
		{
			Path:   "foo",
			Weight: uint32(1),
		},
		{
			Path:   "foo2",
			Weight: uint32(2),
		},
	}
	runtimeThrottleDevicesSlice = []*runtime.ThrottleDevice{
		{
			Path: "foo",
			Rate: uint64(1000),
		},
		{
			Path: "foo2",
			Rate: uint64(2000),
		},
	}
	runtimeUlimitsSlice = []*runtime.Ulimit{
		{
			Name: "foo",
			Hard: 1,
			Soft: 1,
		},
		{
			Name: "foo2",
			Hard: 2,
			Soft: 2,
		},
		{
			Name: "foo3",
			Hard: -1,
			Soft: -1,
		},
	}
	resources = apitypes.Resources{
		CPUPeriod:          1000,
		CPUQuota:           1000,
		CPUShares:          1000,
		Memory:             1000,
		CpusetCpus:         "0",
		CpusetMems:         "0",
		BlkioWeight:        uint16(100),
		BlkioWeightDevice:  apitypesWeightDevicesSlice,
		BlkioDeviceReadBps: apitypesThrottleDevicesSlice,
		MemorySwappiness:   &memorySwappiness,
		Ulimits:            apitypesUlimitsSlice,
	}
	linuxContainerResources = runtime.LinuxContainerResources{
		CpuPeriod:          1000,
		CpuQuota:           1000,
		CpuShares:          1000,
		MemoryLimitInBytes: 1000,
		CpusetCpus:         "0",
		CpusetMems:         "0",
		BlkioWeight:        uint32(100),
		DiskQuota:          map[string]string{"foo": "foo"},
		BlkioWeightDevice:  runtimeWeightDevicesSlice,
		BlkioDeviceReadBps: runtimeThrottleDevicesSlice,
		MemorySwappiness:   &runtime.Int64Value{Value: 1000},
		Ulimits:            runtimeUlimitsSlice,
	}
)

func Test_parseUint32(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    uint32
		wantErr bool
	}{
		{
			name:    "parseTestOk",
			args:    args{s: "123456"},
			want:    uint32(123456),
			wantErr: false,
		},
		{
			name:    "parseTestWrong",
			args:    args{s: "abc"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseUint32(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseUint32() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseUint32() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toCriTimestamp(t *testing.T) {
	type args struct {
		t string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "criTimestampNil",
			args:    args{t: ""},
			want:    0,
			wantErr: false,
		},
		{
			name:    "criTimestampOk",
			args:    args{t: "2018-01-12T07:38:32.245589846Z"},
			want:    int64(1515742712245589846),
			wantErr: false,
		},
		{
			name:    "criTimestampWrongFormat",
			args:    args{t: "abc"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toCriTimestamp(tt.args.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("toCriTimestamp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("toCriTimestamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_generateEnvList(t *testing.T) {
	type args struct {
		envs []*runtime.KeyValue
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
	}{
		{
			name:       "envList",
			args:       args{envs: []*runtime.KeyValue{{Key: "a", Value: "b"}, {Key: "c", Value: "d"}}},
			wantResult: []string{"a=b", "c=d"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotResult := generateEnvList(tt.args.envs); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("generateEnvList() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func Test_makeLabels(t *testing.T) {
	type args struct {
		labels      map[string]string
		annotations map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "makeLabelOnlyLabels",
			args: args{
				labels:      map[string]string{"a": "b", "c": "d"},
				annotations: nil,
			},
			want: map[string]string{"a": "b", "c": "d"},
		},
		{
			name: "makeLabelOnlyAnnotations",
			args: args{
				labels:      nil,
				annotations: map[string]string{"aa": "bb", "cc": "dd"},
			},
			want: map[string]string{annotationPrefix + "aa": "bb", annotationPrefix + "cc": "dd"},
		},
		{
			name: "makeLabelMixed",
			args: args{
				labels:      map[string]string{"a": "b", "c": "d"},
				annotations: map[string]string{"aa": "bb", "cc": "dd"},
			},
			want: map[string]string{"a": "b", "c": "d", annotationPrefix + "aa": "bb", annotationPrefix + "cc": "dd"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeLabels(tt.args.labels, tt.args.annotations); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_extractLabels(t *testing.T) {
	type args struct {
		input map[string]string
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]string
		want1 map[string]string
	}{
		{
			name:  "extractLabelsInternalKey",
			args:  args{input: map[string]string{containerTypeLabelKey: "b", sandboxIDLabelKey: "d"}},
			want:  map[string]string{},
			want1: map[string]string{},
		},
		{
			name:  "extractLabelsOnlyLabels",
			args:  args{input: map[string]string{"aa": "bb", "cc": "dd"}},
			want:  map[string]string{"aa": "bb", "cc": "dd"},
			want1: map[string]string{},
		},
		{
			name:  "extractLabelsOnlyAnnotations",
			args:  args{input: map[string]string{annotationPrefix + "aaa": "bbb", annotationPrefix + "ccc": "ddd"}},
			want:  map[string]string{},
			want1: map[string]string{"aaa": "bbb", "ccc": "ddd"},
		},
		{
			name: "extractLabelsMixed",
			args: args{input: map[string]string{containerTypeLabelKey: "b", sandboxIDLabelKey: "d",
				"aa": "bb", "cc": "dd", annotationPrefix + "aaa": "bbb", annotationPrefix + "ccc": "ddd"}},
			want:  map[string]string{"aa": "bb", "cc": "dd"},
			want1: map[string]string{"aaa": "bbb", "ccc": "ddd"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := extractLabels(tt.args.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractLabels() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("extractLabels() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

// Sandbox related unit tests.

func Test_makeSandboxName(t *testing.T) {
	type args struct {
		c *runtime.PodSandboxConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "sandboxName",
			args: args{
				c: &runtime.PodSandboxConfig{
					Metadata: &runtime.PodSandboxMetadata{Name: "PodSandbox", Namespace: "a", Uid: "e2f34", Attempt: uint32(3)}},
			},
			want: kubePrefix + nameDelimiter + sandboxContainerName + nameDelimiter + "PodSandbox" + nameDelimiter + "a" + nameDelimiter + "e2f34" + nameDelimiter + fmt.Sprintf("%d", uint32(3)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeSandboxName(tt.args.c); got != tt.want {
				t.Errorf("makeSandboxName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseSandboxName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *runtime.PodSandboxMetadata
		wantErr bool
	}{
		{
			name:    "sandboxName",
			args:    args{kubePrefix + nameDelimiter + sandboxContainerName + nameDelimiter + "PodSandbox" + nameDelimiter + "a" + nameDelimiter + "e2f34" + nameDelimiter + fmt.Sprintf("%d", uint32(3))},
			want:    &runtime.PodSandboxMetadata{Name: "PodSandbox", Namespace: "a", Uid: "e2f34", Attempt: uint32(3)},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSandboxName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSandboxName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseSandboxName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toCriSandboxState(t *testing.T) {
	type args struct {
		status apitypes.Status
	}
	tests := []struct {
		name string
		args args
		want runtime.PodSandboxState
	}{
		{
			name: "criSandboxStateReady",
			args: args{status: apitypes.StatusRunning},
			want: runtime.PodSandboxState_SANDBOX_READY,
		},
		{
			name: "criSandboxStateNotReady1",
			args: args{status: apitypes.StatusRestarting},
			want: runtime.PodSandboxState_SANDBOX_NOTREADY,
		},
		{
			name: "criSandboxStateNotReady2",
			args: args{status: apitypes.StatusCreated},
			want: runtime.PodSandboxState_SANDBOX_NOTREADY,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toCriSandboxState(tt.args.status); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toCriSandboxState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_filterCRISandboxes(t *testing.T) {
	testSandboxes := []*runtime.PodSandbox{
		{
			Id:       "1",
			Metadata: &runtime.PodSandboxMetadata{Name: "name-1", Attempt: 1},
			State:    runtime.PodSandboxState_SANDBOX_READY,
			Labels:   map[string]string{"a": "b"},
		},
		{
			Id:       "2",
			Metadata: &runtime.PodSandboxMetadata{Name: "name-2", Attempt: 2},
			State:    runtime.PodSandboxState_SANDBOX_NOTREADY,
			Labels:   map[string]string{"c": "d"},
		},
		{
			Id:       "2",
			Metadata: &runtime.PodSandboxMetadata{Name: "name-3", Attempt: 3},
			State:    runtime.PodSandboxState_SANDBOX_NOTREADY,
			Labels:   map[string]string{"e": "f"},
		},
	}
	for desc, test := range map[string]struct {
		filter *runtime.PodSandboxFilter
		expect []*runtime.PodSandbox
	}{
		"no filter": {
			expect: testSandboxes,
		},
		"id filter": {
			filter: &runtime.PodSandboxFilter{Id: "2"},
			expect: []*runtime.PodSandbox{testSandboxes[1], testSandboxes[2]},
		},
		"state filter": {
			filter: &runtime.PodSandboxFilter{
				State: &runtime.PodSandboxStateValue{
					State: runtime.PodSandboxState_SANDBOX_READY,
				},
			},
			expect: []*runtime.PodSandbox{testSandboxes[0]},
		},
		"label filter": {
			filter: &runtime.PodSandboxFilter{
				LabelSelector: map[string]string{"e": "f"},
			},
			expect: []*runtime.PodSandbox{testSandboxes[2]},
		},
		"mixed filter not matched": {
			filter: &runtime.PodSandboxFilter{
				State: &runtime.PodSandboxStateValue{
					State: runtime.PodSandboxState_SANDBOX_NOTREADY,
				},
				LabelSelector: map[string]string{"a": "b"},
			},
			expect: []*runtime.PodSandbox{},
		},
		"mixed filter matched": {
			filter: &runtime.PodSandboxFilter{
				State: &runtime.PodSandboxStateValue{
					State: runtime.PodSandboxState_SANDBOX_NOTREADY,
				},
				LabelSelector: map[string]string{"c": "d"},
				Id:            "2",
			},
			expect: []*runtime.PodSandbox{testSandboxes[1]},
		},
	} {
		filtered := filterCRISandboxes(testSandboxes, test.filter)
		assert.Equal(t, test.expect, filtered, desc)
	}
}

// Container related unit tests.
func Test_parseContainerName(t *testing.T) {
	format := fmt.Sprintf("%s_${container name}_${sandbox name}_${sandbox namespace}_${sandbox uid}_${attempt times}", kubePrefix)

	longerNameContainer := "k8s_cname_name_namespace_uid_3_4"
	wrongPrefixContainer := "swarm_cname_name_namespace_uid_3"

	wrongAttemptContainer := "k8s_cname_name_namespace_uid_notInt"
	parts := strings.Split(wrongAttemptContainer, nameDelimiter)
	_, wrongAttemptErr := parseUint32(parts[5])

	testCases := []struct {
		input         string
		expectedError string
	}{
		{input: longerNameContainer, expectedError: fmt.Sprintf("failed to parse container name: %q, which should be %s", longerNameContainer, format)},
		{input: wrongPrefixContainer, expectedError: fmt.Sprintf("container is not managed by kubernetes: %q", wrongPrefixContainer)},
		{input: wrongAttemptContainer, expectedError: fmt.Sprintf("failed to parse the attempt times in container name: %q: %v", wrongAttemptContainer, wrongAttemptErr)},
	}

	for _, test := range testCases {
		_, actualError := parseContainerName(test.input)
		assert.EqualError(t, actualError, test.expectedError)
	}
}

func Test_makeupLogPath(t *testing.T) {
	testCases := []struct {
		logDirectory  string
		containerMeta *runtime.ContainerMetadata
		expected      string
	}{
		{
			logDirectory:  "/var/log/pods/099f1c2b79126109140a1f77e211df00",
			containerMeta: &runtime.ContainerMetadata{Name: "kube-scheduler", Attempt: 0},
			expected:      "/var/log/pods/099f1c2b79126109140a1f77e211df00/kube-scheduler/0.log",
		},
		{
			logDirectory:  "/var/log/pods/d875aada-9920-11e8-bfef-0242ac11001e/",
			containerMeta: &runtime.ContainerMetadata{Name: "kube-proxy", Attempt: 10},
			expected:      "/var/log/pods/d875aada-9920-11e8-bfef-0242ac11001e/kube-proxy/10.log",
		},
	}

	for _, test := range testCases {
		logPath := makeupLogPath(test.logDirectory, test.containerMeta)
		if !reflect.DeepEqual(test.expected, logPath) {
			t.Fatalf("unexpected logPath returned by makeupLogPath")
		}
	}
}

func Test_toCriContainerState(t *testing.T) {
	testCases := []struct {
		input    apitypes.Status
		expected runtime.ContainerState
	}{
		{input: apitypes.StatusRunning, expected: runtime.ContainerState_CONTAINER_RUNNING},
		{input: apitypes.StatusExited, expected: runtime.ContainerState_CONTAINER_EXITED},
		{input: apitypes.StatusCreated, expected: runtime.ContainerState_CONTAINER_CREATED},
		{input: apitypes.StatusPaused, expected: runtime.ContainerState_CONTAINER_UNKNOWN},
	}

	for _, test := range testCases {
		actual := toCriContainerState(test.input)
		assert.Equal(t, test.expected, actual)
	}
}

func Test_filterCRIContainers(t *testing.T) {
	testContainers := []*runtime.Container{
		{
			Id:           "1",
			PodSandboxId: "s-1",
			Metadata:     &runtime.ContainerMetadata{Name: "name-1", Attempt: 1},
			State:        runtime.ContainerState_CONTAINER_RUNNING,
		},
		{
			Id:           "2",
			PodSandboxId: "s-2",
			Metadata:     &runtime.ContainerMetadata{Name: "name-2", Attempt: 2},
			State:        runtime.ContainerState_CONTAINER_EXITED,
			Labels:       map[string]string{"a": "b"},
		},
		{
			Id:           "3",
			PodSandboxId: "s-2",
			Metadata:     &runtime.ContainerMetadata{Name: "name-2", Attempt: 3},
			State:        runtime.ContainerState_CONTAINER_CREATED,
			Labels:       map[string]string{"c": "d"},
		},
	}
	for desc, test := range map[string]struct {
		filter *runtime.ContainerFilter
		expect []*runtime.Container
	}{
		"no filter": {
			expect: testContainers,
		},
		"id filter": {
			filter: &runtime.ContainerFilter{Id: "2"},
			expect: []*runtime.Container{testContainers[1]},
		},
		"state filter": {
			filter: &runtime.ContainerFilter{
				State: &runtime.ContainerStateValue{
					State: runtime.ContainerState_CONTAINER_EXITED,
				},
			},
			expect: []*runtime.Container{testContainers[1]},
		},
		"label filter": {
			filter: &runtime.ContainerFilter{
				LabelSelector: map[string]string{"a": "b"},
			},
			expect: []*runtime.Container{testContainers[1]},
		},
		"sandbox id filter": {
			filter: &runtime.ContainerFilter{PodSandboxId: "s-2"},
			expect: []*runtime.Container{testContainers[1], testContainers[2]},
		},
		"mixed filter not matched": {
			filter: &runtime.ContainerFilter{
				Id:            "1",
				PodSandboxId:  "s-2",
				LabelSelector: map[string]string{"a": "b"},
			},
			expect: []*runtime.Container{},
		},
		"mixed filter matched": {
			filter: &runtime.ContainerFilter{
				PodSandboxId: "s-2",
				State: &runtime.ContainerStateValue{
					State: runtime.ContainerState_CONTAINER_CREATED,
				},
				LabelSelector: map[string]string{"c": "d"},
			},
			expect: []*runtime.Container{testContainers[2]},
		},
	} {
		filtered := filterCRIContainers(testContainers, test.filter)
		assert.Equal(t, test.expect, filtered, desc)
	}
}

func Test_modifyContainerNamespaceOptions(t *testing.T) {
	type args struct {
		nsOpts       *runtime.NamespaceOption
		podSandboxID string
		hostConfig   *apitypes.HostConfig
	}
	tests := []struct {
		name string
		args args
		want apitypes.HostConfig
	}{
		{
			name: "normal test",
			args: args{
				nsOpts:       &runtime.NamespaceOption{Network: runtime.NamespaceMode_NODE, Pid: runtime.NamespaceMode_NODE},
				podSandboxID: "fakeSandBoxID",
				hostConfig:   &apitypes.HostConfig{PidMode: "host", IpcMode: "host", NetworkMode: "host"},
			},
			want: apitypes.HostConfig{PidMode: "host", IpcMode: "container:fakeSandBoxID", NetworkMode: "host"},
		},
		{
			name: "nil test",
			args: args{
				nsOpts:       nil,
				podSandboxID: "fakeSandBoxID",
				hostConfig:   &apitypes.HostConfig{PidMode: "host", IpcMode: "host", NetworkMode: "host"},
			},
			want: apitypes.HostConfig{PidMode: "container:fakeSandBoxID", IpcMode: "container:fakeSandBoxID", NetworkMode: "container:fakeSandBoxID"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifyContainerNamespaceOptions(tt.args.nsOpts, tt.args.podSandboxID, tt.args.hostConfig)
			if !reflect.DeepEqual(*tt.args.hostConfig, tt.want) {
				t.Errorf("modifyContainerNamespaceOptions() = %v, want %v", *tt.args.hostConfig, tt.want)
			}
		})
	}
}

func Test_modifyHostConfig(t *testing.T) {
	supplementalGroups := []int64{1, 2, 3}
	groupAdd := []string{}
	for _, group := range supplementalGroups {
		groupAdd = append(groupAdd, strconv.FormatInt(group, 10))
	}

	type args struct {
		sc         *runtime.LinuxContainerSecurityContext
		hostConfig *apitypes.HostConfig
	}
	tests := []struct {
		name           string
		args           args
		wantHostConfig *apitypes.HostConfig
		wantErr        error
	}{
		{
			name: "Normal Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					SupplementalGroups: supplementalGroups,
					Privileged:         true,
					ReadonlyRootfs:     true,
					Capabilities: &runtime.Capability{
						AddCapabilities:  []string{"fooAdd1", "fooAdd2"},
						DropCapabilities: []string{"fooDrop1", "fooDrop2"},
					},
					SeccompProfilePath: mgr.ProfileDockerDefault,
					ApparmorProfile:    mgr.ProfileRuntimeDefault,
					NoNewPrivs:         true,
				},
				hostConfig: &apitypes.HostConfig{},
			},
			wantHostConfig: &apitypes.HostConfig{
				GroupAdd:       groupAdd,
				Privileged:     true,
				ReadonlyRootfs: true,
				CapAdd:         []string{"fooAdd1", "fooAdd2"},
				CapDrop:        []string{"fooDrop1", "fooDrop2"},
				SecurityOpt:    []string{"no-new-privileges"},
			},
			wantErr: nil,
		},
		{
			name: "SupplementalGroups Nil Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					Privileged:     true,
					ReadonlyRootfs: true,
					Capabilities: &runtime.Capability{
						AddCapabilities:  []string{"fooAdd1", "fooAdd2"},
						DropCapabilities: []string{"fooDrop1", "fooDrop2"},
					},
					SeccompProfilePath: mgr.ProfileDockerDefault,
					ApparmorProfile:    mgr.ProfileRuntimeDefault,
					NoNewPrivs:         true,
				},
				hostConfig: &apitypes.HostConfig{},
			},
			wantHostConfig: &apitypes.HostConfig{
				Privileged:     true,
				ReadonlyRootfs: true,
				CapAdd:         []string{"fooAdd1", "fooAdd2"},
				CapDrop:        []string{"fooDrop1", "fooDrop2"},
				SecurityOpt:    []string{"no-new-privileges"},
			},
			wantErr: nil,
		},
		{
			name: "Capabilities Nil Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					SupplementalGroups: supplementalGroups,
					Privileged:         true,
					ReadonlyRootfs:     true,
					SeccompProfilePath: mgr.ProfileDockerDefault,
					ApparmorProfile:    mgr.ProfileRuntimeDefault,
					NoNewPrivs:         true,
				},
				hostConfig: &apitypes.HostConfig{},
			},
			wantHostConfig: &apitypes.HostConfig{
				GroupAdd:       groupAdd,
				Privileged:     true,
				ReadonlyRootfs: true,
				SecurityOpt:    []string{"no-new-privileges"},
			},
			wantErr: nil,
		},
		{
			name: "GetSeccompSecurityOpts Err Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					SupplementalGroups: supplementalGroups,
					Privileged:         true,
					ReadonlyRootfs:     true,
					Capabilities: &runtime.Capability{
						AddCapabilities:  []string{"fooAdd1", "fooAdd2"},
						DropCapabilities: []string{"fooDrop1", "fooDrop2"},
					},
					SeccompProfilePath: "foo",
					ApparmorProfile:    mgr.ProfileRuntimeDefault,
					NoNewPrivs:         true,
				},
				hostConfig: &apitypes.HostConfig{},
			},
			wantHostConfig: &apitypes.HostConfig{
				GroupAdd:       groupAdd,
				Privileged:     true,
				ReadonlyRootfs: true,
				CapAdd:         []string{"fooAdd1", "fooAdd2"},
				CapDrop:        []string{"fooDrop1", "fooDrop2"},
			},
			wantErr: fmt.Errorf("failed to generate seccomp security options: %v", fmt.Errorf("undefault profile %q should prefix with %q", "foo", mgr.ProfileNamePrefix)),
		},
		{
			name: "GetAppArmorSecurityOpts Err Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					SupplementalGroups: supplementalGroups,
					Privileged:         true,
					ReadonlyRootfs:     true,
					Capabilities: &runtime.Capability{
						AddCapabilities:  []string{"fooAdd1", "fooAdd2"},
						DropCapabilities: []string{"fooDrop1", "fooDrop2"},
					},
					SeccompProfilePath: mgr.ProfileDockerDefault,
					ApparmorProfile:    "foo",
					NoNewPrivs:         true,
				},
				hostConfig: &apitypes.HostConfig{},
			},
			wantHostConfig: &apitypes.HostConfig{
				GroupAdd:       groupAdd,
				Privileged:     true,
				ReadonlyRootfs: true,
				CapAdd:         []string{"fooAdd1", "fooAdd2"},
				CapDrop:        []string{"fooDrop1", "fooDrop2"},
			},
			wantErr: fmt.Errorf("failed to generate appArmor security options: %v", fmt.Errorf("undefault profile name should prefix with %q", mgr.ProfileNamePrefix)),
		},
		{
			name: "NoNewPrivs False Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					SupplementalGroups: supplementalGroups,
					Privileged:         true,
					ReadonlyRootfs:     true,
					Capabilities: &runtime.Capability{
						AddCapabilities:  []string{"fooAdd1", "fooAdd2"},
						DropCapabilities: []string{"fooDrop1", "fooDrop2"},
					},
					SeccompProfilePath: mgr.ProfileDockerDefault,
					ApparmorProfile:    mgr.ProfileRuntimeDefault,
					NoNewPrivs:         false,
				},
				hostConfig: &apitypes.HostConfig{},
			},
			wantHostConfig: &apitypes.HostConfig{
				GroupAdd:       groupAdd,
				Privileged:     true,
				ReadonlyRootfs: true,
				CapAdd:         []string{"fooAdd1", "fooAdd2"},
				CapDrop:        []string{"fooDrop1", "fooDrop2"},
			},
			wantErr: nil,
		},
		{
			name: "Nil Test",
			args: args{
				hostConfig: &apitypes.HostConfig{},
			},
			wantHostConfig: &apitypes.HostConfig{},
			wantErr:        nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := modifyHostConfig(tt.args.sc, tt.args.hostConfig)
			if !reflect.DeepEqual(tt.args.hostConfig, tt.wantHostConfig) {
				t.Errorf("modifyHostConfig() hostConfig = %v, wantHostConfig %v", tt.args.hostConfig, tt.wantHostConfig)
				return
			}
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("modifyHostConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_modifyContainerConfig(t *testing.T) {
	runAsUser := &runtime.Int64Value{Value: int64(1)}
	runAsGroup := &runtime.Int64Value{Value: int64(1)}
	formatResult := strconv.FormatInt(1, 10)

	type args struct {
		sc     *runtime.LinuxContainerSecurityContext
		config *apitypes.ContainerConfig
	}
	tests := []struct {
		name       string
		args       args
		wantConfig *apitypes.ContainerConfig
		wantErr    bool
	}{
		{
			name: "No nil Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					RunAsUser:     runAsUser,
					RunAsUsername: "foo",
					RunAsGroup:    runAsGroup,
				},
				config: &apitypes.ContainerConfig{},
			},
			wantConfig: &apitypes.ContainerConfig{
				User: "foo" + ":" + formatResult,
			},
			wantErr: false,
		},
		{
			name: "RunAsUser Nil Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					RunAsUsername: "foo",
					RunAsGroup:    runAsGroup,
				},
				config: &apitypes.ContainerConfig{},
			},
			wantConfig: &apitypes.ContainerConfig{
				User: "foo" + ":" + formatResult,
			},
			wantErr: false,
		},
		{
			name: "RunAsUsername Empty Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					RunAsUser:     runAsUser,
					RunAsUsername: "",
					RunAsGroup:    runAsGroup,
				},
				config: &apitypes.ContainerConfig{},
			},
			wantConfig: &apitypes.ContainerConfig{
				User: formatResult + ":" + formatResult,
			},
			wantErr: false,
		},
		{
			name: "RunAsUser And RunAsUsername All Empty Test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					RunAsUsername: "",
					RunAsGroup:    runAsGroup,
				},
				config: &apitypes.ContainerConfig{},
			},
			wantConfig: &apitypes.ContainerConfig{},
			wantErr:    true,
		},
		{
			name: "Nil Test",
			args: args{
				config: &apitypes.ContainerConfig{},
			},
			wantConfig: &apitypes.ContainerConfig{},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := modifyContainerConfig(tt.args.sc, tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("modifyContainerConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.config, tt.wantConfig) {
				t.Errorf("modifyContainerConfig() config = %v, wantConfig %v", tt.args.config, tt.wantConfig)
				return
			}
		})
	}
}

func Test_toCriContainer(t *testing.T) {
	_, timeParseErr := time.Parse(utils.TimeLayout, "foo")
	type args struct {
		c *mgr.Container
	}
	tests := []struct {
		name    string
		args    args
		want    *runtime.Container
		wantErr error
	}{
		{
			name: "Normal Test",
			args: args{
				c: &mgr.Container{
					ID: "cid",
					State: &apitypes.ContainerState{
						Status: apitypes.StatusRunning,
					},
					Image: "imageRef",
					Name:  "k8s_cname_sname_namespace_uid_3",
					Config: &apitypes.ContainerConfig{
						Image: "image",
						Labels: map[string]string{
							containerTypeLabelKey: "b",
							sandboxIDLabelKey:     "sid",
							"aa":                  "bb",
							"cc":                  "dd",
							annotationPrefix + "aaa": "bbb",
							annotationPrefix + "ccc": "ddd",
						},
					},
					Created: "2018-01-12T07:38:32.245589846Z",
				},
			},
			want: &runtime.Container{
				Id:           "cid",
				PodSandboxId: "sid",
				Metadata: &runtime.ContainerMetadata{
					Name:    "cname",
					Attempt: uint32(3),
				},
				Image:     &runtime.ImageSpec{Image: "image"},
				ImageRef:  "imageRef",
				State:     runtime.ContainerState_CONTAINER_RUNNING,
				CreatedAt: int64(1515742712245589846),
				Labels: map[string]string{
					"aa": "bb",
					"cc": "dd",
				},
				Annotations: map[string]string{
					"aaa": "bbb",
					"ccc": "ddd",
				},
			},
			wantErr: nil,
		},
		{
			name: "ParseContainerName Error Test",
			args: args{
				c: &mgr.Container{
					ID: "cid",
					State: &apitypes.ContainerState{
						Status: apitypes.StatusRunning,
					},
					Image: "imageRef",
					Name:  "kubernetes_cname_sname_namespace_uid_3",
					Config: &apitypes.ContainerConfig{
						Image: "image",
						Labels: map[string]string{
							containerTypeLabelKey: "b",
							sandboxIDLabelKey:     "sid",
							"aa":                  "bb",
							"cc":                  "dd",
							annotationPrefix + "aaa": "bbb",
							annotationPrefix + "ccc": "ddd",
						},
					},
					Created: "2018-01-12T07:38:32.245589846Z",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("container is not managed by kubernetes: %q", "kubernetes_cname_sname_namespace_uid_3"),
		},
		{
			name: "ToCriTimestamp Error Test",
			args: args{
				c: &mgr.Container{
					ID: "cid",
					State: &apitypes.ContainerState{
						Status: apitypes.StatusRunning,
					},
					Image: "imageRef",
					Name:  "k8s_cname_sname_namespace_uid_3",
					Config: &apitypes.ContainerConfig{
						Image: "image",
						Labels: map[string]string{
							containerTypeLabelKey: "b",
							sandboxIDLabelKey:     "sid",
							"aa":                  "bb",
							"cc":                  "dd",
							annotationPrefix + "aaa": "bbb",
							annotationPrefix + "ccc": "ddd",
						},
					},
					Created: "foo",
				},
			},
			want:    nil,
			wantErr: fmt.Errorf("failed to parse create timestamp for container %q: %v", "cid", timeParseErr),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toCriContainer(tt.args.c)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("toCriContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toCriContainer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_containerNetns(t *testing.T) {
	type args struct {
		container *mgr.Container
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Normal Test",
			args: args{
				container: &mgr.Container{
					State: &apitypes.ContainerState{
						Pid: int64(1001),
					},
				},
			},
			want: fmt.Sprintf("/proc/%v/ns/net", 1001),
		},
		{
			name: "Pid EQ 0 Test",
			args: args{
				container: &mgr.Container{
					State: &apitypes.ContainerState{
						Pid: int64(0),
					},
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containerNetns(tt.args.container); got != tt.want {
				t.Errorf("containerNetns() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Image related unit tests.
func Test_imageToCriImage(t *testing.T) {
	repoDigests := []string{"lastest", "dev", "v1.0"}
	imageUserInt := "1"
	uid, _ := strconv.ParseInt(imageUserInt, 10, 64)
	containerVolumes := map[string]interface{}{"foo": "foo"}
	runtimeVolumes := parseVolumesFromPouch(containerVolumes)

	type args struct {
		image *apitypes.ImageInfo
	}
	tests := []struct {
		name    string
		args    args
		want    *runtime.Image
		wantErr error
	}{
		{
			name: "Normal Test",
			args: args{
				image: &apitypes.ImageInfo{
					ID:          "image-id",
					RepoTags:    repoDigests,
					RepoDigests: repoDigests,
					Size:        1024,
					Config: &apitypes.ContainerConfig{
						User:    imageUserInt,
						Volumes: containerVolumes,
					},
				},
			},
			want: &runtime.Image{
				Id:          "image-id",
				RepoTags:    repoDigests,
				RepoDigests: repoDigests,
				Size_:       uint64(1024),
				Uid:         &runtime.Int64Value{Value: uid},
				Username:    "",
				Volumes:     runtimeVolumes,
			},
			wantErr: nil,
		},
		{
			name: "ImageUID Nil Test",
			args: args{
				image: &apitypes.ImageInfo{
					ID:          "image-id",
					RepoTags:    repoDigests,
					RepoDigests: repoDigests,
					Size:        1024,
					Config: &apitypes.ContainerConfig{
						User:    "foo",
						Volumes: containerVolumes,
					},
				},
			},
			want: &runtime.Image{
				Id:          "image-id",
				RepoTags:    repoDigests,
				RepoDigests: repoDigests,
				Size_:       uint64(1024),
				Uid:         nil,
				Username:    "foo",
				Volumes:     runtimeVolumes,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := imageToCriImage(tt.args.image)
			if err != tt.wantErr {
				t.Errorf("imageToCriImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("imageToCriImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getUserFromImageUser(t *testing.T) {
	imageUserInt := "1"
	uid, _ := strconv.ParseInt(imageUserInt, 10, 64)
	type args struct {
		imageUser string
	}
	tests := []struct {
		name         string
		args         args
		wantUID      *int64
		wantUserName string
	}{
		{
			name: "Empty Test",
			args: args{
				imageUser: "",
			},
			wantUID:      nil,
			wantUserName: "",
		},
		{
			name: "ParseInt Success Test",
			args: args{
				imageUser: imageUserInt,
			},
			wantUID:      &uid,
			wantUserName: "",
		},
		{
			name: "ParseInt Fail Test",
			args: args{
				imageUser: "foo",
			},
			wantUID:      nil,
			wantUserName: "foo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUID, gotUsername := getUserFromImageUser(tt.args.imageUser)
			if (gotUID == nil && tt.wantUID != nil) || (gotUID != nil && tt.wantUID == nil) {
				t.Errorf("getUserFromImageUser() gotUID = %v, wantUID %v", gotUID, tt.wantUID)
			}
			if gotUID != nil && tt.wantUID != nil {
				if (*gotUID) != (*tt.wantUID) {
					t.Errorf("getUserFromImageUser() gotUID = %v, wantUID %v", gotUID, tt.wantUID)
				}
			}
			if gotUsername != tt.wantUserName {
				t.Errorf("getUserFromImageUser() gotUsername = %v, wantUserName %v", gotUsername, tt.wantUserName)
			}
		})
	}
}

func Test_parseUserFromImageUser(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Empty Test",
			args: args{
				id: "",
			},
			want: "",
		},
		{
			name: "user:group Test",
			args: args{
				id: "user:group",
			},
			want: "user",
		},
		{
			name: "No Group Test",
			args: args{
				id: "user",
			},
			want: "user",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseUserFromImageUser(tt.args.id); got != tt.want {
				t.Errorf("parseUserFromImageUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

// CRI extension related tool functions tests.

func Test_parseResourcesFromCRI(t *testing.T) {
	type args struct {
		runtimeResources *runtime.LinuxContainerResources
	}
	tests := []struct {
		name string
		args args
		want apitypes.Resources
	}{
		{
			name: "normal test",
			args: args{
				runtimeResources: &linuxContainerResources,
			},
			want: resources,
		},
		{
			name: "nil test",
			args: args{
				runtimeResources: &runtime.LinuxContainerResources{},
			},
			want: apitypes.Resources{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseResourcesFromCRI(tt.args.runtimeResources); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseResourcesFromCRI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseResourcesFromPouch(t *testing.T) {
	type args struct {
		apitypesResources apitypes.Resources
		diskQuota         map[string]string
	}
	tests := []struct {
		name string
		args args
		want *runtime.LinuxContainerResources
	}{
		{
			name: "normal test",
			args: args{
				apitypesResources: resources,
				diskQuota:         map[string]string{"foo": "foo"},
			},
			want: &linuxContainerResources,
		},
		{
			name: "nil test",
			args: args{
				apitypesResources: apitypes.Resources{},
				diskQuota:         nil,
			},
			want: &runtime.LinuxContainerResources{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseResourcesFromPouch(tt.args.apitypesResources, tt.args.diskQuota); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseResourcesFromPouch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseWeightDeviceFromCRI(t *testing.T) {
	type args struct {
		runtimeWeightDevices []*runtime.WeightDevice
	}
	tests := []struct {
		name              string
		args              args
		wantWeightDevices []*apitypes.WeightDevice
	}{
		{
			name: "normal test",
			args: args{
				runtimeWeightDevices: runtimeWeightDevicesSlice,
			},
			wantWeightDevices: apitypesWeightDevicesSlice,
		},
		{
			name: "nil test",
			args: args{
				runtimeWeightDevices: nil,
			},
			wantWeightDevices: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotWeightDevices := parseWeightDeviceFromCRI(tt.args.runtimeWeightDevices); !reflect.DeepEqual(gotWeightDevices, tt.wantWeightDevices) {
				t.Errorf("parseWeightDeviceFromCRI() = %v, want %v", gotWeightDevices, tt.wantWeightDevices)
			}
		})
	}
}

func Test_parseWeightDeviceFromPouch(t *testing.T) {
	type args struct {
		apitypesWeightDevices []*apitypes.WeightDevice
	}
	tests := []struct {
		name              string
		args              args
		wantWeightDevices []*runtime.WeightDevice
	}{
		{
			name: "normal test",
			args: args{
				apitypesWeightDevices: apitypesWeightDevicesSlice,
			},
			wantWeightDevices: runtimeWeightDevicesSlice,
		},
		{
			name: "nil test",
			args: args{
				apitypesWeightDevices: nil,
			},
			wantWeightDevices: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotWeightDevices := parseWeightDeviceFromPouch(tt.args.apitypesWeightDevices); !reflect.DeepEqual(gotWeightDevices, tt.wantWeightDevices) {
				t.Errorf("parseWeightDeviceFromPouch() = %v, want %v", gotWeightDevices, tt.wantWeightDevices)
			}
		})
	}
}

func Test_parseThrottleDeviceFromCRI(t *testing.T) {
	type args struct {
		runtimeThrottleDevices []*runtime.ThrottleDevice
	}
	tests := []struct {
		name                string
		args                args
		wantThrottleDevices []*apitypes.ThrottleDevice
	}{
		{
			name: "normal test",
			args: args{
				runtimeThrottleDevices: runtimeThrottleDevicesSlice,
			},
			wantThrottleDevices: apitypesThrottleDevicesSlice,
		},
		{
			name: "nil test",
			args: args{
				runtimeThrottleDevices: nil,
			},
			wantThrottleDevices: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotThrottleDevices := parseThrottleDeviceFromCRI(tt.args.runtimeThrottleDevices); !reflect.DeepEqual(gotThrottleDevices, tt.wantThrottleDevices) {
				t.Errorf("parseThrottleDeviceFromCRI() = %v, want %v", gotThrottleDevices, tt.wantThrottleDevices)
			}
		})
	}
}

func Test_parseThrottleDeviceFromPouch(t *testing.T) {
	type args struct {
		apitypesThrottleDevices []*apitypes.ThrottleDevice
	}
	tests := []struct {
		name                string
		args                args
		wantThrottleDevices []*runtime.ThrottleDevice
	}{
		{
			name: "normal test",
			args: args{
				apitypesThrottleDevices: apitypesThrottleDevicesSlice,
			},
			wantThrottleDevices: runtimeThrottleDevicesSlice,
		},
		{
			name: "nil test",
			args: args{
				apitypesThrottleDevices: nil,
			},
			wantThrottleDevices: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotThrottleDevices := parseThrottleDeviceFromPouch(tt.args.apitypesThrottleDevices); !reflect.DeepEqual(gotThrottleDevices, tt.wantThrottleDevices) {
				t.Errorf("parseThrottleDeviceFromPouch() = %v, want %v", gotThrottleDevices, tt.wantThrottleDevices)
			}
		})
	}
}

func Test_parseUlimitFromCRI(t *testing.T) {
	type args struct {
		runtimeUlimits []*runtime.Ulimit
	}
	tests := []struct {
		name        string
		args        args
		wantUlimits []*apitypes.Ulimit
	}{
		{
			name: "normal test",
			args: args{
				runtimeUlimits: runtimeUlimitsSlice,
			},
			wantUlimits: apitypesUlimitsSlice,
		},
		{
			name: "nil test",
			args: args{
				runtimeUlimits: nil,
			},
			wantUlimits: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUlimits := parseUlimitFromCRI(tt.args.runtimeUlimits); !reflect.DeepEqual(gotUlimits, tt.wantUlimits) {
				t.Errorf("parseUlimitFromCRI() = %v, want %v", gotUlimits, tt.wantUlimits)
			}
		})
	}
}

func Test_parseUlimitFromPouch(t *testing.T) {
	type args struct {
		apitypesUlimits []*apitypes.Ulimit
	}
	tests := []struct {
		name        string
		args        args
		wantUlimits []*runtime.Ulimit
	}{
		{
			name: "normal test",
			args: args{
				apitypesUlimits: apitypesUlimitsSlice,
			},
			wantUlimits: runtimeUlimitsSlice,
		},
		{
			name: "nil test",
			args: args{
				apitypesUlimits: nil,
			},
			wantUlimits: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotUlimits := parseUlimitFromPouch(tt.args.apitypesUlimits); !reflect.DeepEqual(gotUlimits, tt.wantUlimits) {
				t.Errorf("parseUlimitFromPouch() = %v, want %v", gotUlimits, tt.wantUlimits)
			}
		})
	}
}

func Test_parseVolumesFromPouch(t *testing.T) {
	type args struct {
		containerVolumes map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]*runtime.Volume
	}{
		{
			name: "normal test",
			args: args{
				containerVolumes: map[string]interface{}{"foo": "foo"},
			},
			want: map[string]*runtime.Volume{"foo": {}},
		},
		{
			name: "nil test",
			args: args{
				containerVolumes: make(map[string]interface{}),
			},
			want: make(map[string]*runtime.Volume),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseVolumesFromPouch(tt.args.containerVolumes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseVolumesFromPouch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseEnvsFromPouch(t *testing.T) {
	type args struct {
		pouchEnvs []string
	}
	tests := []struct {
		name        string
		args        args
		wantCriEnvs []*runtime.KeyValue
	}{
		{
			name: "Normal Test",
			args: args{
				pouchEnvs: []string{"key1=value1", "key2=value2"},
			},
			wantCriEnvs: []*runtime.KeyValue{
				{
					Key:   "key1",
					Value: "value1",
				},
				{
					Key:   "key2",
					Value: "value2",
				},
			},
		},
		{
			name: "Multiple Equals Sign Test",
			args: args{
				pouchEnvs: []string{"key1=value1", "key2=foo=value2"},
			},
			wantCriEnvs: []*runtime.KeyValue{
				{
					Key:   "key1",
					Value: "value1",
				},
				{
					Key:   "key2",
					Value: "foo=value2",
				},
			},
		},
		{
			name: "No Equals Sign Test",
			args: args{
				pouchEnvs: []string{"key1=value1", "key2"},
			},
			wantCriEnvs: []*runtime.KeyValue{
				{
					Key:   "key1",
					Value: "value1",
				},
				{
					Key: "key2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotCriEnvs := parseEnvsFromPouch(tt.args.pouchEnvs); !reflect.DeepEqual(gotCriEnvs, tt.wantCriEnvs) {
				t.Errorf("parseEnvsFromPouch() = %v, want %v", gotCriEnvs, tt.wantCriEnvs)
			}
		})
	}
}

func Test_generateMountBindings(t *testing.T) {
	type args struct {
		mounts []*runtime.Mount
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "propagation_private test",
			args: args{
				mounts: []*runtime.Mount{
					{
						ContainerPath:  "container_path",
						HostPath:       "host_path",
						Readonly:       true,
						SelinuxRelabel: true,
						Propagation:    runtime.MountPropagation_PROPAGATION_PRIVATE,
					},
				},
			},
			want: []string{"host_path:container_path:ro,Z"},
		},
		{
			name: "propagation_bidirectinal test",
			args: args{
				mounts: []*runtime.Mount{
					{
						ContainerPath:  "container_path",
						HostPath:       "host_path",
						Readonly:       true,
						SelinuxRelabel: false,
						Propagation:    runtime.MountPropagation_PROPAGATION_BIDIRECTIONAL,
					},
				},
			},
			want: []string{"host_path:container_path:ro,rshared"},
		},
		{
			name: "propagation_host_to_container test",
			args: args{
				mounts: []*runtime.Mount{
					{
						ContainerPath:  "container_path",
						HostPath:       "host_path",
						Readonly:       false,
						SelinuxRelabel: true,
						Propagation:    runtime.MountPropagation_PROPAGATION_HOST_TO_CONTAINER,
					},
				},
			},
			want: []string{"host_path:container_path:Z,rslave"},
		},
		{
			name: "no_attrs test",
			args: args{
				mounts: []*runtime.Mount{
					{
						ContainerPath:  "container_path",
						HostPath:       "host_path",
						Readonly:       false,
						SelinuxRelabel: false,
					},
				},
			},
			want: []string{"host_path:container_path"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := generateMountBindings(tt.args.mounts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("generateMountBindings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_modifySandboxNamespaceOptions(t *testing.T) {
	type args struct {
		nsOpts     *runtime.NamespaceOption
		hostConfig *apitypes.HostConfig
	}
	tests := []struct {
		name string
		args args
		want *apitypes.HostConfig
	}{
		{
			name: "nil test",
			args: args{
				nsOpts: &runtime.NamespaceOption{},
				hostConfig: &apitypes.HostConfig{
					IpcMode:     namespaceModeHost,
					PidMode:     namespaceModeHost,
					NetworkMode: namespaceModeHost,
				},
			},
			want: &apitypes.HostConfig{
				IpcMode:     namespaceModeHost,
				PidMode:     namespaceModeHost,
				NetworkMode: namespaceModeHost,
			},
		},
		{
			name: "normal test",
			args: args{
				nsOpts: &runtime.NamespaceOption{
					Ipc: runtime.NamespaceMode_NODE,
					Pid: runtime.NamespaceMode_NODE,
				},
				hostConfig: &apitypes.HostConfig{},
			},
			want: &apitypes.HostConfig{
				IpcMode: namespaceModeHost,
				PidMode: namespaceModeHost,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifySandboxNamespaceOptions(tt.args.nsOpts, tt.args.hostConfig)
			if !reflect.DeepEqual(tt.args.hostConfig, tt.want) {
				t.Errorf("modifySandboxNamespaceOptions() = %v, want %v", tt.args.hostConfig, tt.want)
			}
		})
	}
}

// CNI Network related unit tests.
func Test_toCNIPortMappings(t *testing.T) {
	criNormalTCP := &runtime.PortMapping{
		Protocol:      runtime.Protocol_TCP,
		ContainerPort: 8080,
		HostPort:      80,
		HostIp:        "192.168.1.101",
	}
	pouchNormalTCP := ocicni.PortMapping{
		Protocol:      "tcp",
		ContainerPort: 8080,
		HostPort:      80,
		HostIP:        "192.168.1.101",
	}
	criNormalUDP := &runtime.PortMapping{
		Protocol:      runtime.Protocol_UDP,
		ContainerPort: 8080,
		HostPort:      80,
		HostIp:        "192.168.1.102",
	}
	pouchNormalUDP := ocicni.PortMapping{
		Protocol:      "udp",
		ContainerPort: 8080,
		HostPort:      80,
		HostIP:        "192.168.1.102",
	}
	criHostPortLEZero := &runtime.PortMapping{
		Protocol:      runtime.Protocol_TCP,
		ContainerPort: 8080,
		HostPort:      0,
		HostIp:        "192.168.1.100",
	}

	type args struct {
		criPortMappings []*runtime.PortMapping
	}
	tests := []struct {
		name string
		args args
		want []ocicni.PortMapping
	}{
		{
			name: "Normal Test",
			args: args{
				criPortMappings: []*runtime.PortMapping{
					criNormalTCP,
					criNormalUDP,
				},
			},
			want: []ocicni.PortMapping{
				pouchNormalTCP,
				pouchNormalUDP,
			},
		},
		{
			name: "HostPort LE Zero Test",
			args: args{
				criPortMappings: []*runtime.PortMapping{
					criNormalTCP,
					criNormalUDP,
					criHostPortLEZero,
				},
			},
			want: []ocicni.PortMapping{
				pouchNormalTCP,
				pouchNormalUDP,
			},
		},
		{
			name: "Nil Test",
			args: args{
				criPortMappings: []*runtime.PortMapping{},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toCNIPortMappings(tt.args.criPortMappings); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toCNIPortMappings() = %v, want %v", got, tt.want)
			}
		})
	}
}
