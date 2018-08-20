package user

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
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

// uidParser defines lines in /etc/passwd, eg: root:x:0:0:root:/root:/bin/bash.
type uidParser struct {
	user        string
	placeholder string
	uid         int
	gid         int
	finger      []string
	userdir     string
	shell       string
}

// gidParser defines lines in /etc/group, eg: root:x:0:.
type gidParser struct {
	group       string
	placeholder string
	gid         int
	otherGroup  []string
}

// Get accepts user and group slice, return valid uid, gid and additional gids.
// Through Get is a interface returns all user informations runtime-spec need,
// GetUser, GetIntegerID, GetAdditionalGids still can be used independently.
func Get(passwdPath, groupPath, user string, groups []string) (uint32, uint32, []uint32, error) {
	uid, gid, err := GetUser(passwdPath, groupPath, user)
	if err != nil {
		return 0, 0, nil, err
	}

	return uid, gid, GetAdditionalGids(groups), nil
}

// GetUser accepts user string like <uid|username>:<gid|groupname>, and transfers them to format valid uid:gid.
// user format example:
// user
// uid
// uid:gid
// user:group
// uid:group
// user:gid
func GetUser(passwdPath, groupPath, user string) (uint32, uint32, error) {
	if user == "" {
		// if user is null, return 0 value as root user
		return 0, 0, nil
	}

	var (
		uidStr, gidStr string
		uid, gid       uint32
		err            error
	)

	ParseString(user, &uidStr, &gidStr)

	// get uid from /etc/passwd
	uid, err = ParseID(filepath.Join(passwdPath, PasswdFile), uidStr, func(line, str string, idInt int, idErr error) (uint32, bool) {
		var up uidParser
		ParseString(line, &up.user, &up.placeholder, &up.uid)
		if (idErr == nil && idInt == up.uid) || str == up.user {
			return uint32(up.uid), true
		}
		return 0, false
	})
	if err != nil {
		// if uidStr is a integer, treat it as valid uid
		integer, e := strconv.Atoi(uidStr)
		if e != nil {
			return 0, 0, err
		}
		uid = uint32(integer)
	}

	// if gidStr is null, then get gid from /etc/passwd
	if len(gidStr) == 0 {
		gid, err = ParseID(filepath.Join(passwdPath, PasswdFile), uidStr, func(line, str string, idInt int, idErr error) (uint32, bool) {
			var up uidParser
			ParseString(line, &up.user, &up.placeholder, &up.uid, &up.gid)
			if (idErr == nil && idInt == up.uid) || str == up.user {
				return uint32(up.gid), true
			}
			return 0, false
		})
	} else {
		gid, err = ParseID(filepath.Join(groupPath, GroupFile), gidStr, func(line, str string, idInt int, idErr error) (uint32, bool) {
			var gp gidParser
			ParseString(line, &gp.group, &gp.placeholder, &gp.gid)
			if (idErr == nil && idInt == gp.gid) || str == gp.group {
				return uint32(gp.gid), true
			}
			return 0, false
		})
	}
	if err != nil {
		// if gidStr is a integer, treat it as valid gid
		integer, e := strconv.Atoi(gidStr)
		if e != nil {
			return 0, 0, err
		}
		gid = uint32(integer)
	}

	return uid, gid, nil
}

// GetIntegerID only parser user format uid:gid, cause container rootfs is not created
// by contianerd now, can not change user to id, only support user id >= 1000
// TODO(huamin.thm): removed later
func GetIntegerID(user string) (uint32, uint32) {
	if user == "" {
		// return default user root
		return 0, 0
	}

	// if uid gid can not be parsed successfully, return default user root
	var uid, gid int
	ParseString(user, &uid, &gid)
	return uint32(uid), uint32(gid)
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
