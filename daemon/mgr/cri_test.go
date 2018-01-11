package mgr

import (
	"testing"

	apitypes "github.com/alibaba/pouch/apis/types"
	"github.com/stretchr/testify/assert"

	"k8s.io/kubernetes/pkg/kubelet/apis/cri/v1alpha1/runtime"
)

func TestParseUint32(t *testing.T) {
	testCases := []struct {
		input    string
		expected uint32
	}{
		{input: "0", expected: uint32(0)},
		{input: "1", expected: uint32(1)},
	}

	for _, test := range testCases {
		actual, err := parseUint32(test.input)
		assert.NoError(t, err)
		assert.Equal(t, test.expected, actual)
	}
}

func TestLabelsAndAnnotationsRoundTrip(t *testing.T) {
	expectedLabels := map[string]string{"label.123.abc": "foo", "label.456.efg": "bar"}
	expectedAnnotations := map[string]string{"annotation.abc.123": "uvw", "annotation.def.456": "xyz"}
	// Merge labels and annotations into pouch labels.
	pouchLabels := makeLabels(expectedLabels, expectedAnnotations)
	// Extract labels and annotations from pouch labels.
	actualLabels, actualAnnotations := extractLabels(pouchLabels)
	assert.Equal(t, expectedLabels, actualLabels)
	assert.Equal(t, expectedAnnotations, actualAnnotations)
}

// Sandbox related tests.

func makeSandboxConfigWithLabelsAndAnnotations(name, namespace, uid string, attempt uint32, labels, annotations map[string]string) *runtime.PodSandboxConfig {
	return &runtime.PodSandboxConfig{
		Metadata: &runtime.PodSandboxMetadata{
			Name:      name,
			Namespace: namespace,
			Uid:       uid,
			Attempt:   attempt,
		},
		Labels:      labels,
		Annotations: annotations,
	}
}

// A helper to create a basic config.
func makeSandboxConfig(name, namespace, uid string, attempt uint32) *runtime.PodSandboxConfig {
	return makeSandboxConfigWithLabelsAndAnnotations(name, namespace, uid, attempt, map[string]string{}, map[string]string{})
}

func TestSandboxNameRoundTrip(t *testing.T) {
	config := makeSandboxConfig("name", "namespace", "uid", 1)
	actualName := makeSandboxName(config)
	assert.Equal(t, "k8s_POD_name_namespace_uid_1", actualName)

	actualMetadata, err := parseSandboxName(actualName)
	assert.NoError(t, err)
	assert.Equal(t, config.Metadata, actualMetadata)
}

func TestToCriSandboxState(t *testing.T) {
	testCases := []struct {
		input    apitypes.Status
		expected runtime.PodSandboxState
	}{
		{input: apitypes.StatusRunning, expected: runtime.PodSandboxState_SANDBOX_READY},
		{input: apitypes.StatusExited, expected: runtime.PodSandboxState_SANDBOX_NOTREADY},
	}

	for _, test := range testCases {
		actual := toCriSandboxState(test.input)
		assert.Equal(t, test.expected, actual)
	}
}

// Container related unit tests.

func TestContainerNameRoundTrip(t *testing.T) {
	sandboxConfig := makeSandboxConfig("name", "namespace", "uid", 1)
	name, attempt := "cname", uint32(3)
	config := &runtime.ContainerConfig{
		Metadata: &runtime.ContainerMetadata{
			Name:    name,
			Attempt: attempt,
		},
	}
	actualName := makeContainerName(sandboxConfig, config)
	assert.Equal(t, "k8s_cname_name_namespace_uid_3", actualName)

	actualMetadata, err := parseContainerName(actualName)
	assert.NoError(t, err)
	assert.Equal(t, config.Metadata, actualMetadata)
}

func TestToCriContainerState(t *testing.T) {
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
