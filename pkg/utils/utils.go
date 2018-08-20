package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

// If implements ternary operator. if cond is true return v1, or return v2 instead.
func If(cond bool, v1, v2 interface{}) interface{} {
	if cond {
		return v1
	}
	return v2
}

// FormatSize format image size to B/KB/MB/GB
func FormatSize(size int64) string {
	if size <= 0 {
		return "0.00 B"
	}
	// we consider image size less than 1024 GB
	suffixes := []string{"B", "KB", "MB", "GB"}

	var count int
	formattedSize := float64(size)
	for count = 0; count < 3; count++ {
		if formattedSize < 1024 {
			break
		}
		formattedSize /= 1024
	}

	return fmt.Sprintf("%.2f %s", formattedSize, suffixes[count])
}

// TruncateID is used to transfer image ID from digest to short ID.
func TruncateID(id string) string {
	var shortLen = 12

	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > shortLen {
		return id[:shortLen]
	}
	return id
}

// Merge merge object from src to dest, dest object should be pointer, only accept struct type, notice: src will overwrite dest's data
func Merge(src, dest interface{}) error {
	if src == nil || dest == nil {
		return fmt.Errorf("merged object can not be nil")
	}

	destType := reflect.TypeOf(dest)
	if destType.Kind() != reflect.Ptr {
		return fmt.Errorf("merged object not pointer")
	}
	destVal := reflect.ValueOf(dest).Elem()

	if destVal.Kind() != reflect.Struct {
		return fmt.Errorf("merged object type should be struct")
	}

	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}
	if destVal.Type() != srcVal.Type() {
		return fmt.Errorf("src and dest object type must same")
	}

	return doMerge(srcVal, destVal)
}

// doMerge, begin merge action, note that we will merge slice type,
// but we do not validate if slice has duplicate values.
func doMerge(src, dest reflect.Value) error {
	if !src.IsValid() || !dest.CanSet() || isEmptyValue(src) {
		return nil
	}

	switch dest.Kind() {
	case reflect.Struct:
		for i := 0; i < dest.NumField(); i++ {
			if err := doMerge(src.Field(i), dest.Field(i)); err != nil {
				return err
			}
		}

	case reflect.Map:
		for _, key := range src.MapKeys() {
			srcElem := src.MapIndex(key)
			if !srcElem.IsValid() || isEmptyValue(srcElem) {
				continue
			}
			if dest.IsNil() {
				dest.Set(reflect.MakeMap(dest.Type()))
			}
			dest.SetMapIndex(key, srcElem)
		}

	case reflect.Slice:
		dest.Set(reflect.AppendSlice(dest, src))

	default:
		dest.Set(src)
	}

	return nil
}

// From src/pkg/encoding/json,
// we recognize nullable values like `false` `0` as not empty.
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// DeDuplicate make a slice with no duplicated elements.
func DeDuplicate(input []string) []string {
	if input == nil {
		return nil
	}
	result := []string{}
	internal := map[string]struct{}{}
	for _, value := range input {
		if _, exist := internal[value]; !exist {
			internal[value] = struct{}{}
			result = append(result, value)
		}
	}
	return result
}

// FormatErrMsgFunc is a function which used by CombineErrors to
// format error message
type FormatErrMsgFunc func(idx int, err error) (string, error)

// CombineErrors is a function which used by Inspect to merge multiple errors
// into one error.
func CombineErrors(errs []error, formatErrMsg FormatErrMsgFunc) error {
	var errMsgs []string
	for idx, err := range errs {
		formattedErrMsg, formatError := formatErrMsg(idx, err)
		if formatError != nil {
			return fmt.Errorf("Combine errors error: %s", formatError.Error())
		}
		errMsgs = append(errMsgs, formattedErrMsg)
	}
	combinedErrMsg := strings.Join(errMsgs, "\n")
	return errors.New(combinedErrMsg)
}

// Contains check if a interface in a interface slice.
func Contains(input []interface{}, value interface{}) (bool, error) {
	if value == nil || len(input) == 0 {
		return false, nil
	}

	if reflect.TypeOf(input[0]) != reflect.TypeOf(value) {
		return false, fmt.Errorf("interface type not equals")
	}

	switch v := value.(type) {
	case int, int64, float64, string:
		for _, v := range input {
			if v == value {
				return true, nil
			}
		}
		return false, nil
	// TODO: add more types
	default:
		r := reflect.TypeOf(v)
		return false, fmt.Errorf("Not support: %s", r)
	}
}

// StringInSlice checks if a string in the slice.
func StringInSlice(input []string, str string) bool {
	if str == "" || len(input) == 0 {
		return false
	}

	result := make([]interface{}, len(input))
	for i, v := range input {
		result[i] = v
	}

	exists, _ := Contains(result, str)
	return exists
}

// checkPidfileStatus check if pidfile exist and validate pid exist in /proc, but not validate whether process is running.
func checkPidfileStatus(path string) error {
	if pidByte, err := ioutil.ReadFile(path); err == nil {
		if _, err := os.Stat("/proc/" + string(pidByte)); err == nil {
			return fmt.Errorf("found daemon pid %s, check it status", string(pidByte))
		}
	}

	return nil
}

// NewPidfile checks if pidfile exist, and saves daemon pid.
func NewPidfile(path string) error {
	if err := checkPidfileStatus(path); err != nil {
		return err
	}
	return ioutil.WriteFile(path, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

// IsProcessAlive returns true if process with a given pid is running.
func IsProcessAlive(pid int) bool {
	err := syscall.Kill(pid, syscall.Signal(0))
	if err == nil || err == syscall.EPERM {
		return true
	}

	return false
}

// KillProcess force-stops a process.
func KillProcess(pid int) {
	syscall.Kill(pid, syscall.SIGKILL)
}

// SetOOMScore sets process's oom_score value
// The higher the value of oom_score of any process, the higher is its
// likelihood of getting killed by the OOM Killer in an out-of-memory situation.
func SetOOMScore(pid, score int) error {
	f, err := os.OpenFile(fmt.Sprintf("/proc/%d/oom_score_adj", pid), os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	_, err = f.WriteString(strconv.Itoa(score))
	f.Close()
	return err
}

// ConvertKVStringsToMap converts ["key=value"] into {"key":"value"}
func ConvertKVStringsToMap(values []string) (map[string]string, error) {
	kvs := make(map[string]string, len(values))

	for _, value := range values {
		terms := strings.SplitN(value, "=", 2)
		if len(terms) != 2 {
			return nil, fmt.Errorf("input %s must have format of key=value", value)
		}
		kvs[terms[0]] = terms[1]
	}
	return kvs, nil
}

// ConvertKVStrToMapWithNoErr converts input strings and converts them all in a map,
// When there is invalid input, the dealing procedure ignores the error and log a warning message.
func ConvertKVStrToMapWithNoErr(values []string) map[string]string {
	kvs := make(map[string]string, len(values))
	for _, value := range values {
		k, v, err := ConvertStrToKV(value)
		if err != nil {
			logrus.Warnf("input %s should have a format of key=value", value)
			continue
		}
		kvs[k] = v
	}
	return kvs
}

// ConvertStrToKV converts an string into key and value string without returning an error.
// For example, for input "a=b", it should return "a", "b".
func ConvertStrToKV(input string) (string, string, error) {
	results := strings.SplitN(input, "=", 2)
	if len(results) != 2 {
		return "", "", fmt.Errorf("input string %s must have format key=value", input)
	}
	return results[0], results[1], nil
}

// IsFileExist checks if file is exits on host.
func IsFileExist(file string) bool {
	if _, err := os.Stat(file); err == nil {
		return true
	}

	return false
}

// StringSliceEqual compare two string slice, ignore the order.
func StringSliceEqual(s1, s2 []string) bool {
	if s1 == nil && s2 == nil {
		return true
	}

	if s1 == nil || s2 == nil {
		return false
	}

	if len(s1) != len(s2) {
		return false
	}

	for _, s := range s1 {
		if !StringInSlice(s2, s) {
			return false
		}
	}

	for _, s := range s2 {
		if !StringInSlice(s1, s) {
			return false
		}
	}

	return true
}
