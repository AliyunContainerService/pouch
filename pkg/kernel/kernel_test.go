package kernel

import (
	"regexp"
	"testing"

	"github.com/alibaba/pouch/pkg/exec"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

var (
	// regex for specific characters in  kernel version
	kernelVersionReg = regexp.MustCompile(`[^0-9A-Za-z\._-]`)
)

func TestGetKernelVersion(t *testing.T) {
	version, err := GetKernelVersion()
	assert.Equal(t, nil, err)

	println(version.String())

	if kernelVersionReg.MatchString(version.Flavor) {
		t.Errorf("kernel version's flavor has unexpected specific characters: %+v", []byte(version.Flavor))
	}
}

// Benchmark result for below two methods to execute `uname` command in GetKernelVersion().

// BenchmarkGetUnameByUnix-4      	  200000	     10584 ns/op
// BenchmarkGetUnameByExecRun-4   	     200	   6255530 ns/op

// Benchmark for executing `uname` by raw unix system call
func BenchmarkGetUnameByUnix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buf := unix.Utsname{}
		unix.Uname(&buf)
	}
}

// Benchmark for executing `uname` by original method -- clone and run the command.
func BenchmarkGetUnameByExecRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		exec.Run(0, "uname", "-r")
	}
}
