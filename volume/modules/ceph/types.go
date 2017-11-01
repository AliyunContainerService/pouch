// +build linux

package ceph

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/alibaba/pouch/pkg/exec"
)

const (
	// HealthOK denotes the status of ceph cluster when healthy.
	HealthOK = "HEALTH_OK"

	// HealthWarn denotes the status of ceph cluster when unhealthy but recovering.
	HealthWarn = "HEALTH_WARN"

	// HealthErr denotes the status of ceph cluster when unhealthy but usually needs
	// manual intervention.
	HealthErr = "HEALTH_ERR"
)

var (
	// QuorumCommand represents ceph quorum_status command.
	QuorumCommand = []string{"quorum_status", "-f", "json"}

	// CephUsageCommand represents ceph df command.
	CephUsageCommand = []string{"df", "detail", "-f", "json"}

	// CephStatusCommand represents ceph status command.
	CephStatusCommand = []string{"status", "-f", "json"}

	// OsdDFCommand represents ceph osd df command.
	OsdDFCommand = []string{"osd", "df", "-f", "json"}

	// OsdPerfCommand represents ceph osd perf command.
	OsdPerfCommand = []string{"osd", "perf", "-f", "json"}

	// OsdDumpCommand represents ceph osd dump command.
	OsdDumpCommand = []string{"osd", "dump", "-f", "json"}
)

var (
	defaultTimeout       = time.Second * 10
	defaultFormatTimeout = time.Second * 120
)

type mon struct {
	Addr string `json:"addr"`
	Name string `json:"name"`
	Rank int    `json:"rank"`
}

type monMap struct {
	Created  string `json:"created"`
	Epoch    int    `json:"epoch"`
	FsID     string `json:"fsid"`
	Modified string `json:"modified"`
	Mons     []mon  `json:"mons"`
}

// QuorumStatus represents ceph quorum_status result struct.
type QuorumStatus struct {
	MonMap           monMap `json:"monmap"`
	QuorumLeaderName string `json:"quorum_leader_name"`
}

// Stats represents ceph status result struct.
type Stats struct {
	ID     string `json:"fsid"`
	Health struct {
		OverallStatus string `json:"overall_status"`
	} `json:"health"`
}

// PoolStats represents ceph pool status struct.
type PoolStats struct {
	Pools []struct {
		Name  string `json:"name"`
		Stats struct {
			BytesUsed int64 `json:"bytes_used"`
			MaxAvail  int64 `json:"max_avail"`
		} `json:"stats"`
	} `json:"pools"`
}

// RBDMap represents "rbd listmapped" command result struct.
type RBDMap struct {
	MapDevice []struct {
		ID     int    `json:"id"`
		Device string `json:"name"`
		Image  string `json:"image"`
		Pool   string `json:"pool"`
	} `json:"Devices"`
}

// Command represents ceph command.
type Command struct {
	bin  string
	conf string
}

// NewCephCommand returns ceph command instance.
func NewCephCommand(bin, conf string) *Command {
	return &Command{bin, conf}
}

// RunCommand is used to execute ceph command.
func (cc *Command) RunCommand(obj interface{}, args ...string) error {
	args = append(args, "--conf")
	args = append(args, cc.conf)

	exit, stdout, stderr, err := exec.Run(defaultTimeout, cc.bin, args...)
	if err != nil || exit != 0 {
		return fmt.Errorf("run command failed, err: %v, exit: %d, output: %s", err, exit, stderr)
	}

	if obj != nil {
		if err := json.Unmarshal([]byte(stdout), obj); err != nil {
			return err
		}
	}

	return nil
}
