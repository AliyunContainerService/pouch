// +build linux

package local

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alibaba/pouch/pkg/exec"
	"github.com/sirupsen/logrus"
)

type unitMap map[string]int64

const (
	// KB represents kilo bytes.
	KB = 1024

	// MB represents mega bytes.
	MB = 1024 * KB

	//GB represents giga bytes.
	GB = 1024 * MB

	// TB represents tera bytes.
	TB = 1024 * GB

	// PB represents peta bytes.
	PB = 1024 * TB
)

var (
	lock        sync.Mutex
	mountPoints = make(map[uint64]string)
	quotaClient = http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
		Timeout: time.Second * 30,
	}
	decimalMap = unitMap{"k": KB, "m": MB, "g": GB, "t": TB, "p": PB}
	sizeRegex  = regexp.MustCompile(`^(\d+)([kKmMgGtTpP])?[bB]?$`)
)

func quotaDriverStart(dir string) (string, error) {
	devID, err := getDevID(dir)
	if err != nil {
		return "", err
	}

	lock.Lock()
	defer lock.Unlock()

	if mp, ok := mountPoints[devID]; ok {
		return mp, nil
	}

	mountPoint, hasQuota, _ := checkMountpoint(devID)
	if len(mountPoint) == 0 {
		return mountPoint, fmt.Errorf("mountPoint not found: %s", dir)
	}
	if !hasQuota {
		_, _, _, err := exec.Run(0, "mount", "-o", "remount,grpquota", mountPoint)
		if err != nil {
			return "", err
		}
	}

	vfsVersion, quotaFilename, err := getVFSVersionAndQuotaFile(devID)
	if err != nil {
		return "", err
	}

	filename := mountPoint + "/" + quotaFilename
	if _, err := os.Stat(filename); err != nil {
		os.Remove(mountPoint + "/aquota.user")

		header := []byte{0x27, 0x19, 0xc0, 0xd9, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3a, 0x09, 0x00, 0x80,
			0x3a, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x05, 0x00, 0x00, 0x00}
		if vfsVersion == "vfsv1" {
			header[4] = 0x01
		}

		if writeErr := ioutil.WriteFile(filename, header, 644); writeErr != nil {
			logrus.Errorf("write file error. %s, %s, %s", filename, vfsVersion, writeErr)
			return mountPoint, writeErr
		}

		if _, _, _, err := exec.Run(0, "setquota", "-g", "-t", "43200", "43200", mountPoint); err != nil {
			os.Remove(filename)
			return mountPoint, err
		}
		if err := setUserQuota(0, 0, mountPoint); err != nil {
			os.Remove(filename)
			return mountPoint, err
		}
	}

	// on
	exit, stdout, stderr, err := exec.Run(0, "quotaon", "-pg", mountPoint)
	if err != nil && exit != 1 {
		logrus.Errorf("quotaon failed, exit: %d, stdout: %s, stderr: %s, err: %v", exit, stdout, stderr, err)
		return "", fmt.Errorf("stderr: %s, err: %v", stderr, err)
	}
	if strings.Contains(stdout, " is on") {
		mountPoints[devID] = mountPoint
		return mountPoint, nil
	}
	if _, _, _, err = exec.Run(0, "quotaon", mountPoint); err != nil {
		mountPoint = ""
	}

	mountPoints[devID] = mountPoint
	return mountPoint, err
}

//setfattr -n system.subtree -v $QUOTAID
func setSubtree(dir string, qid uint32) (uint32, error) {
	id := qid
	var err error
	if id == 0 {
		id = getFileAttr(dir)
		if id > 0 {
			return id, nil
		}
		id, err = getNextQuatoID()
	}

	if err != nil {
		return 0, err
	}
	strid := strconv.FormatUint(uint64(id), 10)
	_, _, _, err = exec.Run(0, "setfattr", "-n", "system.subtree", "-v", strid, dir)
	return id, err
}

func setDiskQuota(dir string, size string, quotaID int) error {
	mountPoint, err := quotaDriverStart(dir)
	if err != nil {
		return err
	}
	if len(mountPoint) == 0 {
		return fmt.Errorf("mountpoint not found: %s", dir)
	}

	id, err := setSubtree(dir, uint32(quotaID))
	if id == 0 {
		return fmt.Errorf("subtree not found: %s %v", dir, err)
	}

	limit, _ := fromHumanSize(size)
	if limit == 0 {
		logrus.Infof("size is zero for dir %s, no need to set disk quota", dir)
		return nil
	}
	limit = limit / KB
	if limit <= 0 {
		limit = 1
	}
	return setUserQuota(id, uint64(limit), mountPoint)
}

func getDevID(dir string) (uint64, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(dir, &st); err != nil {
		logrus.Warnf("getDirDev: %s, %v", dir, err)
		return 0, err
	}
	return st.Dev, nil
}

func checkMountpoint(devID uint64) (string, bool, string) {
	output, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		logrus.Warnf("ReadFile: %v", err)
		return "", false, ""
	}

	var mountPoint, fsType string
	hasQuota := false
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, " ")
		if len(parts) != 6 {
			continue
		}
		devID2, _ := getDevID(parts[1])
		if devID == devID2 {
			mountPoint = parts[1]
			fsType = parts[2]
			for _, opt := range strings.Split(parts[3], ",") {
				if opt == "grpquota" {
					hasQuota = true
				}
			}
			break
		}
	}

	return mountPoint, hasQuota, fsType
}

func setUserQuota(quotaID uint32, diskQuota uint64, mountPoint string) error {
	uid := strconv.FormatUint(uint64(quotaID), 10)
	limit := strconv.FormatUint(diskQuota, 10)
	_, _, _, err := exec.Run(0, "setquota", "-g", uid, "0", limit, "0", "0", mountPoint)
	return err
}

//getfattr -n system.subtree --only-values --absolute-names /
func getFileAttr(dir string) uint32 {
	v := 0
	_, out, _, err := exec.Run(0, "getfattr", "-n", "system.subtree", "--only-values", "--absolute-names", dir)
	if err == nil {
		v, _ = strconv.Atoi(out)
	}
	return uint32(v)
}

//next id
func getNextQuatoID() (uint32, error) {
	resp, e := quotaClient.Get("http://127.0.0.1/volume/nextQuotaId")
	if e != nil {
		return 0, e
	}
	defer resp.Body.Close()
	var s struct {
		QuotaID uint32 `json:"quotaId"`
	}
	e = json.NewDecoder(resp.Body).Decode(&s)
	if e != nil {
		return 0, e
	}
	return s.QuotaID, nil
}

func getVFSVersionAndQuotaFile(devID uint64) (string, string, error) {
	output, err := ioutil.ReadFile("/proc/mounts")
	if err != nil {
		logrus.Warnf("ReadFile: %v", err)
		return "", "", err
	}

	vfsVersion := "vfsv0"
	quotaFilename := "aquota.group"
	for _, line := range strings.Split(string(output), "\n") {
		parts := strings.Split(line, " ")
		if len(parts) != 6 {
			continue
		}

		devID2, _ := getDevID(parts[1])
		if devID == devID2 {
			for _, opt := range strings.Split(parts[3], ",") {
				items := strings.SplitN(opt, "=", 2)
				if len(items) != 2 {
					continue
				}
				switch items[0] {
				case "jqfmt":
					vfsVersion = items[1]
				case "grpjquota":
					quotaFilename = items[1]
				}
			}
			break
		}
	}

	return vfsVersion, quotaFilename, nil
}

// FromHumanSize convert human readable size variable to bytes in int64
func fromHumanSize(size string) (int64, error) {
	return parseSize(size, decimalMap)
}

func parseSize(sizeStr string, uMap unitMap) (int64, error) {
	matches := sizeRegex.FindStringSubmatch(sizeStr)
	if len(matches) != 3 {
		return -1, fmt.Errorf("invalid size: '%s'", sizeStr)
	}

	size, err := strconv.ParseInt(matches[1], 10, 0)
	if err != nil {
		return -1, err
	}

	unitPrefix := strings.ToLower(matches[2])
	if mul, ok := uMap[unitPrefix]; ok {
		size *= mul
	}

	return size, nil
}
