package user

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/alibaba/pouch/pkg/log"
	"github.com/opencontainers/runc/libcontainer/user"
)

var (
	// PasswdFile keeps user passwd information
	PasswdFile = "/etc/passwd"
	// GroupFile keeps group information
	GroupFile = "/etc/group"

	minID      = 0
	maxID      = 1<<31 - 1 // compatible for 32-bit OS
	acceptedID = 1000
)

type filterFunc func(line, str string, idInt int, idErr error) (uint32, bool)

// Get accepts user and group slice, return valid uid, gid and additional gids.
// Through Get is a interface returns all user informations runtime-spec need.
func Get(passwdPath, groupPath, username string, groups []string) (uint32, uint32, []uint32, error) {
	log.With(nil).Debugf("get users, passwd path: (%s), group path: (%s), username: (%s), groups: (%v)",
		passwdPath, groupPath, username, groups)

	if passwdPath == "" || groupPath == "" {
		log.With(nil).Warn("get passwd file or group file is nil")
	}

	passwdFile, err := os.Open(passwdPath)
	if err == nil {
		defer passwdFile.Close()
	}

	groupFile, err := os.Open(groupPath)
	if err == nil {
		defer groupFile.Close()
	}

	execUser, err := user.GetExecUser(username, nil, passwdFile, groupFile)
	if err != nil {
		return 0, 0, nil, err
	}

	var addGroups []int
	if len(groups) > 0 {
		addGroups, err = user.GetAdditionalGroups(groups, groupFile)
		if err != nil {
			return 0, 0, nil, err
		}
	}
	uid := uint32(execUser.Uid)
	gid := uint32(execUser.Gid)
	sgids := append(execUser.Sgids, addGroups...)
	var additionalGids []uint32
	for _, g := range sgids {
		additionalGids = append(additionalGids, uint32(g))
	}

	return uid, gid, additionalGids, nil
}

// GetAdditionalGids parse supplementary gids from slice groups.
func GetAdditionalGids(groups []string) []uint32 {
	var additionalGids []uint32

	// TODO: check whether group is valid and support group name format like "nobody".
	for _, group := range groups {
		gid, err := strconv.ParseUint(group, 10, 32)
		if err != nil {
			continue
		}
		additionalGids = append(additionalGids, uint32(gid))
	}

	return additionalGids
}

// ParseID parses id or name from given file.
func ParseID(file, str string, parserFilter filterFunc) (uint32, error) {
	idInt, idErr := strconv.Atoi(str)

	ba, err := ioutil.ReadFile(file)
	if err != nil {
		return 0, fmt.Errorf("failed to read passwd file %s: %s", PasswdFile, err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(ba))
	for scanner.Scan() {
		line := scanner.Text()
		id, ok := parserFilter(line, str, idInt, idErr)
		if ok {
			return id, nil
		}
	}

	if idErr == nil {
		ok, err := isUnknownUser(idInt)
		if err != nil {
			return 0, err
		}
		if ok {
			return uint32(idInt), nil
		}
	}

	return 0, fmt.Errorf("failed to find %s in %s", str, file)
}

// isUnknownUser determines if id can be accepted as a unknown user. this kind of user id should >= 1000
func isUnknownUser(id int) (bool, error) {
	// first test id is valid in minID ~ maxID
	if id < minID || id > maxID {
		return false, fmt.Errorf("use id %d out of range, should be in %d ~ %d", id, minID, maxID)
	}

	if id < acceptedID {
		return false, nil
	}

	return true, nil
}

// ParseString parses line in format a:b:c.
func ParseString(line string, v ...interface{}) {
	splits := strings.Split(line, ":")
	for i, s := range splits {
		if len(v) <= i {
			break
		}

		switch p := v[i].(type) {
		case *string:
			*p = s
		case *int:
			*p, _ = strconv.Atoi(s)
		case *[]string:
			ss := strings.Split(s, ",")
			if len(ss) > 0 {
				*p = ss
			} else {
				*p = []string{}
			}
		}

	}
}
