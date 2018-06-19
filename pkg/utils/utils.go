package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Common durations that is .
// There are some definitions for units of Day and larger .
const (
	Second = time.Second
	Minute = Second * 60
	Hour   = Minute * 60
	Day    = Hour * 24
	Week   = Day * 7
	Month  = Day * 30
	Year   = Day * 365

	TimeLayout = time.RFC3339Nano
)

var errInvalid = errors.New("invalid time")

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

// FormatTimeInterval is used to show the time interval from input time to now.
func FormatTimeInterval(input int64) (formattedTime string, err error) {
	start := time.Unix(0, input)
	diff := time.Now().Sub(start)

	// That should not happen.
	if diff < 0 {
		return "", errInvalid
	}

	timeThresholds := []time.Duration{Year, Month, Week, Day, Hour, Minute, Second}
	timeNames := []string{"year", "month", "week", "day", "hour", "minute", "second"}

	for i, threshold := range timeThresholds {
		if diff >= threshold {
			count := int(diff / threshold)
			formattedTime += strconv.Itoa(count) + " " + timeNames[i]
			if count > 1 {
				formattedTime += "s"
			}
			break
		}
	}

	if diff < Second {
		formattedTime += "0 second"
	}

	return formattedTime, nil
}

// GetUnixTimestamp will parse the value into time and get the nano-timestamp
// in string.
//
// NOTE: if the value is not relative time, GetUnixTimestamp will use RFC3339
// format to parse the value.
func GetUnixTimestamp(value string, base time.Time) (string, error) {
	// time.ParseDuration will handle the 5h, 7d relative time.
	if d, err := time.ParseDuration(value); value != "0" && err == nil {
		return strconv.FormatInt(base.Add(-d).Unix(), 10), nil
	}

	var (
		// rfc3399
		layoutDate            = "2006-01-02"
		layoutDateWithH       = "2006-01-02T15"
		layoutDateWithHM      = "2006-01-02T15:04"
		layoutDateWithHMS     = "2006-01-02T15:04:05"
		layoutDateWithHMSNano = "2006-01-02T15:04:05.999999999"

		layout string
	)

	// if the value doesn't contain any z, Z, +, T, : and -, it maybe
	// timestamp and we should return value.
	if !strings.ContainsAny(value, "zZ+.:T-") {
		return value, nil
	}

	// if the value containns any z, Z or +, we should parse it with timezone
	isLocal := !(strings.ContainsAny(value, "zZ+") || strings.Count(value, "-") == 3)

	if strings.Contains(value, ".") {
		// if the value contains ., we should parse it with nano
		if isLocal {
			layout = layoutDateWithHMSNano
		} else {
			layout = layoutDateWithHMSNano + "Z07:00"
		}
	} else if strings.Contains(value, "T") {
		// if the value contains T, we should parse it with h:m:s
		numColons := strings.Count(value, ":")

		// NOTE:
		// from https://tools.ietf.org/html/rfc3339
		//
		// time-numoffset = ("+" / "-") time-hour [[":"] time-minute]
		//
		// if the value has zero with +/-, it may contains the extra
		// colon like +08:00, which we should remove the extra colon.
		if !isLocal && !strings.ContainsAny(value, "zZ") && numColons > 0 {
			numColons--
		}

		switch numColons {
		case 0:
			layout = layoutDateWithH
		case 1:
			layout = layoutDateWithHM
		default:
			layout = layoutDateWithHMS
		}

		if !isLocal {
			layout += "Z07:00"
		}
	} else if isLocal {
		layout = layoutDate
	} else {
		layout = layoutDate + "Z07:00"
	}

	var t time.Time
	var err error

	if isLocal {
		t, err = time.ParseInLocation(layout, value, time.FixedZone(base.Zone()))
	} else {
		t, err = time.Parse(layout, value)
	}

	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%d.%09d", t.Unix(), int64(t.Nanosecond())), nil
}

// ParseTimestamp returns seconds and nanoseconds.
//
// 1. If the value is empty, it will return default second, the second arg.
// 2. If the incoming nanosecond portion is longer or shorter than 9 digits,
//	it will be converted into 9 digits nanoseconds.
func ParseTimestamp(value string, defaultSec int64) (int64, int64, error) {
	if value == "" {
		return defaultSec, 0, nil
	}

	vs := strings.SplitN(value, ".", 2)

	// for second
	s, err := strconv.ParseInt(vs[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	if len(vs) != 2 {
		return s, 0, nil
	}

	// for nanoseconds
	n, err := strconv.ParseInt(vs[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}

	// convert the n into 9 digits
	n = int64(float64(n) * math.Pow(float64(10), float64(9-len(vs[1]))))
	return s, n, nil
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
			return nil, errors.New("the format must be key=value")
		}
		kvs[terms[0]] = terms[1]
	}
	return kvs, nil
}
