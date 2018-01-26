package mgr

import (
	"reflect"
	"testing"

	apitypes "github.com/alibaba/pouch/apis/types"
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
	// TODO: Add test cases.
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
	// TODO: Add test cases.
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
	// TODO: Add test cases.
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
	// TODO: Add test cases.
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
	// TODO: Add test cases.
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

func Test_makeSandboxName(t *testing.T) {
	type args struct {
		c *runtime.PodSandboxConfig
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
	// TODO: Add test cases.
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
	// TODO: Add test cases.
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
		c *ContainerMeta
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
	type args struct {
		sandboxes []*runtime.PodSandbox
		filter    *runtime.PodSandboxFilter
	}
	tests := []struct {
		name string
		args args
		want []*runtime.PodSandbox
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterCRISandboxes(tt.args.sandboxes, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterCRISandboxes() = %v, want %v", got, tt.want)
			}
		})
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

func Test_parseContainerName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    *runtime.ContainerMetadata
		wantErr bool
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseContainerName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseContainerName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseContainerName() = %v, want %v", got, tt.want)
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
		ContainerMgr ContainerMgr
		ImageMgr     ImageMgr
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

func Test_toCriContainerState(t *testing.T) {
	type args struct {
		status apitypes.Status
	}
	tests := []struct {
		name string
		args args
		want runtime.ContainerState
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toCriContainerState(tt.args.status); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toCriContainerState() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toCriContainer(t *testing.T) {
	type args struct {
		c *ContainerMeta
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

func Test_filterCRIContainers(t *testing.T) {
	type args struct {
		containers []*runtime.Container
		filter     *runtime.ContainerFilter
	}
	tests := []struct {
		name string
		args args
		want []*runtime.Container
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterCRIContainers(tt.args.containers, tt.args.filter); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterCRIContainers() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		ContainerMgr ContainerMgr
		ImageMgr     ImageMgr
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
