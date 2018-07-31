package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/mgr"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
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

func Test_makeSandboxPouchConfig(t *testing.T) {
	type args struct {
		config *runtime.PodSandboxConfig
		image  string
	}
	tests := []struct {
		name    string
		args    args
		want    *apitypes.ContainerCreateConfig
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeSandboxPouchConfig(tt.args.config, tt.args.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("makeSandboxPouchConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeSandboxPouchConfig() = %v, want %v", got, tt.want)
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

func Test_toCriSandbox(t *testing.T) {
	type args struct {
		c *mgr.Container
	}
	tests := []struct {
		name    string
		args    args
		want    *runtime.PodSandbox
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toCriSandbox(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("toCriSandbox() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toCriSandbox() = %v, want %v", got, tt.want)
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

func Test_makeContainerName(t *testing.T) {
	type args struct {
		s *runtime.PodSandboxConfig
		c *runtime.ContainerConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeContainerName(tt.args.s, tt.args.c); got != tt.want {
				t.Errorf("makeContainerName() = %v, want %v", got, tt.want)
			}
		})
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
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifyContainerNamespaceOptions(tt.args.nsOpts, tt.args.podSandboxID, tt.args.hostConfig)
		})
	}
}

func Test_applyContainerSecurityContext(t *testing.T) {
	type args struct {
		lc           *runtime.LinuxContainerConfig
		podSandboxID string
		config       *apitypes.ContainerConfig
		hc           *apitypes.HostConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := applyContainerSecurityContext(tt.args.lc, tt.args.podSandboxID, tt.args.config, tt.args.hc); (err != nil) != tt.wantErr {
				t.Errorf("applyContainerSecurityContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCriManager_updateCreateConfig(t *testing.T) {
	type fields struct {
		ContainerMgr mgr.ContainerMgr
		ImageMgr     mgr.ImageMgr
	}
	type args struct {
		createConfig  *apitypes.ContainerCreateConfig
		config        *runtime.ContainerConfig
		sandboxConfig *runtime.PodSandboxConfig
		podSandboxID  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CriManager{
				ContainerMgr: tt.fields.ContainerMgr,
				ImageMgr:     tt.fields.ImageMgr,
			}
			if err := c.updateCreateConfig(tt.args.createConfig, tt.args.config, tt.args.sandboxConfig, tt.args.podSandboxID); (err != nil) != tt.wantErr {
				t.Errorf("CriManager.updateCreateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_toCriContainer(t *testing.T) {
	type args struct {
		c *mgr.Container
	}
	tests := []struct {
		name    string
		args    args
		want    *runtime.Container
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toCriContainer(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("toCriContainer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toCriContainer() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Image related unit tests.
func Test_imageToCriImage(t *testing.T) {
	type args struct {
		image *apitypes.ImageInfo
	}
	tests := []struct {
		name    string
		args    args
		want    *runtime.Image
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := imageToCriImage(tt.args.image)
			if (err != nil) != tt.wantErr {
				t.Errorf("imageToCriImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("imageToCriImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCriManager_ensureSandboxImageExists(t *testing.T) {
	type fields struct {
		ContainerMgr mgr.ContainerMgr
		ImageMgr     mgr.ImageMgr
	}
	type args struct {
		ctx   context.Context
		image string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CriManager{
				ContainerMgr: tt.fields.ContainerMgr,
				ImageMgr:     tt.fields.ImageMgr,
			}
			if err := c.ensureSandboxImageExists(tt.args.ctx, tt.args.image); (err != nil) != tt.wantErr {
				t.Errorf("CriManager.ensureSandboxImageExists() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_getUserFromImageUser(t *testing.T) {
	type args struct {
		imageUser string
	}
	tests := []struct {
		name  string
		args  args
		want  *int64
		want1 string
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getUserFromImageUser(tt.args.imageUser)
			if got != tt.want {
				t.Errorf("getUserFromImageUser() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getUserFromImageUser() got1 = %v, want %v", got1, tt.want1)
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
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseUserFromImageUser(tt.args.id); got != tt.want {
				t.Errorf("parseUserFromImageUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseResourcesFromCRI(t *testing.T) {
	var (
		resources = apitypes.Resources{
			CPUPeriod:  1000,
			CPUQuota:   1000,
			CPUShares:  1000,
			Memory:     1000,
			CpusetCpus: "0",
			CpusetMems: "0",
		}
		linuxContainerResources = runtime.LinuxContainerResources{
			CpuPeriod:          1000,
			CpuQuota:           1000,
			CpuShares:          1000,
			MemoryLimitInBytes: 1000,
			CpusetCpus:         "0",
			CpusetMems:         "0",
		}
	)
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

func Test_getSELinuxSecurityOpts(t *testing.T) {
	type args struct {
		sc *runtime.LinuxContainerSecurityContext
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "normal test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					SelinuxOptions: &runtime.SELinuxOption{User: "system_u", Role: "object_r", Type: "type", Level: "s0:c123,c456"},
				},
			},
			want: []string{"label=user:system_u", "label=role:object_r", "label=type:type", "label=level:s0:c123,c456"},
		},
		{
			name: "incomplete test",
			args: args{
				sc: &runtime.LinuxContainerSecurityContext{
					SelinuxOptions: &runtime.SELinuxOption{User: "user", Role: "", Type: "type", Level: ""},
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result, err := getSELinuxSecurityOpts(tt.args.sc); err != nil {
				t.Errorf("getSELinuxSecurityOpts() error = %v", err)
			} else if !reflect.DeepEqual(result, tt.want) {
				t.Errorf("getSELinuxSecurityOpts() = %v, want %v", result, tt.want)
			}
		})
	}
}
