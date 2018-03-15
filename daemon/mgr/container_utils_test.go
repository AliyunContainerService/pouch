package mgr

import (
	"path"
	"reflect"
	"testing"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/pkg/collect"
	"github.com/alibaba/pouch/pkg/meta"

	"github.com/stretchr/testify/assert"
)

func TestContainerManager_generateID(t *testing.T) {
	store, err := meta.NewStore(meta.Config{
		Driver:  "local",
		BaseDir: path.Join("/tmp", "containers"),
		Buckets: []meta.Bucket{
			{
				Name: meta.MetaJSONFile,
				Type: reflect.TypeOf(ContainerMeta{}),
			},
		},
	})
	assert.NoError(t, err)

	containerMgr := &ContainerManager{
		NameToID: collect.NewSafeMap(),
		Store:    store,
	}

	id, err := containerMgr.generateID()
	assert.Equal(t, len(id), 64)
	assert.NoError(t, err)
}

func TestContainerManager_generateName(t *testing.T) {
	containerMgr := &ContainerManager{
		NameToID: collect.NewSafeMap(),
	}

	// length less than 6, return empty string
	inputName := "aa"
	generatedName := containerMgr.generateName(inputName)
	assert.Equal(t, generatedName, "")

	inputName = "90719b5f9a455b3314a49e72e3ecb9962f215e0f90153aa8911882acf2ba2c84"
	generatedName = containerMgr.generateName(inputName)
	assert.Equal(t, generatedName, "90719b")

	// store another element which is a prefix of generated ID
	containerMgr.NameToID.Put("90719b", "90719b5f9a455b3314a49e72e3ecb9962f215e0f90153aa8911882acf2ba2c84")
	assert.True(t, containerMgr.NameToID.Get("90719b").Exist())

	// input this name twice
	inputName = "90719b5f9a455b3314a49e72e3ecb9962f215e0f90153aa8911882acf2ba2c84"
	generatedName = containerMgr.generateName(inputName)
	assert.Equal(t, generatedName, "0719b5")

	// store an element previously
	containerMgr.NameToID.Put("aaaaaa", "aaaaaaaaaaaa")
	assert.True(t, containerMgr.NameToID.Get("aaaaaa").Exist())

	inputName = "aaaaaaaaaaaa"
	generatedName = containerMgr.generateName(inputName)
	// according to func generateName, it returns aaaaaa,
	// but this is a bug.
	// FIXME and fix the func generateName
	assert.Equal(t, generatedName, "aaaaaa")

	inputName = "aaaaaaaaaaaab"
	generatedName = containerMgr.generateName(inputName)
	assert.Equal(t, generatedName, "aaaaab")

	inputName = "abcdefghigk"
	generatedName = containerMgr.generateName(inputName)
	assert.Equal(t, generatedName, "abcdef")
}

func Test_parseSecurityOpt(t *testing.T) {
	type args struct {
		meta        *ContainerMeta
		securityOpt string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid security option",
			args: args{
				meta:        &ContainerMeta{},
				securityOpt: "",
			},
			wantErr: true,
		},
		{
			name: "invalid security option",
			args: args{
				meta:        &ContainerMeta{},
				securityOpt: "apparmor:/tmp/file",
			},
			wantErr: true,
		},
		{
			name: "invalid security option",
			args: args{
				meta:        &ContainerMeta{},
				securityOpt: "apparmor2=/tmp/file",
			},
			wantErr: true,
		},
		{
			name: "valid security option",
			args: args{
				meta:        &ContainerMeta{},
				securityOpt: "apparmor=/tmp/file",
			},
			wantErr: false,
		},
		{
			name: "valid security option",
			args: args{
				meta:        &ContainerMeta{},
				securityOpt: "seccomp=asdfghjkl",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := parseSecurityOpt(tt.args.meta, tt.args.securityOpt); (err != nil) != tt.wantErr {
				t.Errorf("parseSecurityOpt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_parsePSOutput(t *testing.T) {
	type args struct {
		output []byte
		pids   []int
	}
	tests := []struct {
		name    string
		args    args
		want    *types.ContainerProcessList
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "testParsePSOutputOk",
			args: args{
				output: []byte("UID        PID  PPID  C STIME TTY          TIME CMD\nroot         1     0  0 3月12 ?       00:00:14 /usr/lib/systemd/systemd --switched-root --system --deserialize 21"),
				pids:   []int{1},
			},
			want: &types.ContainerProcessList{
				Processes: [][]string{
					{"root", "1", "0", "0", "3月12", "?", "00:00:14", "/usr/lib/systemd/systemd --switched-root --system --deserialize 21"},
				},
				Titles: []string{"UID", "PID", "PPID", "C", "STIME", "TTY", "TIME", "CMD"},
			},
			wantErr: false,
		},
		{
			name: "testParsePSOutputWithNoPID",
			args: args{
				output: []byte("UID        PPID  C STIME TTY          TIME CMD\nroot         0  0 3月12 ?       00:00:14 /usr/lib/systemd/systemd --switched-root --system --deserialize 21"),
				pids:   []int{1},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePSOutput(tt.args.output, tt.args.pids)
			if (err != nil) != tt.wantErr {
				t.Errorf("parsePSOutput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePSOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}
