package mgr

import (
	"reflect"
	"testing"
	"time"

	"github.com/alibaba/pouch/apis/types"
	"github.com/alibaba/pouch/daemon/logger"
)

func TestConvContainerLogsOptionToReadConfig(t *testing.T) {
	for _, tc := range []struct {
		input    *types.ContainerLogsOptions
		expected *logger.ReadConfig
		hasError bool
	}{
		{
			input:    &types.ContainerLogsOptions{},
			expected: &logger.ReadConfig{Tail: -1},
			hasError: false,
		}, {
			input: &types.ContainerLogsOptions{
				Since:  "20180510.0001",
				Until:  "20180510.0002",
				Tail:   "100",
				Follow: true,
			},
			expected: &logger.ReadConfig{
				Since:  time.Unix(20180510, 100000),
				Until:  time.Unix(20180510, 200000),
				Tail:   100,
				Follow: true,
			},
			hasError: false,
		}, {
			input: &types.ContainerLogsOptions{
				Since: "20180510.bar",
			},
			expected: nil,
			hasError: true,
		}, {
			input: &types.ContainerLogsOptions{
				Until: "20180510.foo",
			},
			expected: nil,
			hasError: true,
		},
	} {
		got, err := convContainerLogsOptionsToReadConfig(tc.input)
		if err == nil && tc.hasError {
			t.Fatal("expect error, but got nothing")
		}

		if err != nil && !tc.hasError {
			t.Fatalf("expect no error, but got %v", err)
		}

		if !reflect.DeepEqual(got, tc.expected) {
			t.Fatalf("expect %v, but got %v", tc.expected, got)
		}
	}
}
