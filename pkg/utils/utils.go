package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
)

const (
	sigrtmin = 34
	sigrtmax = 64
)

// SignalMap is a map of Linux signals.
var SignalMap = map[string]syscall.Signal{
	"ABRT":     syscall.SIGABRT,
	"ALRM":     syscall.SIGALRM,
	"BUS":      syscall.SIGBUS,
	"CHLD":     syscall.SIGCHLD,
	"CLD":      syscall.SIGCLD,
	"CONT":     syscall.SIGCONT,
	"FPE":      syscall.SIGFPE,
	"HUP":      syscall.SIGHUP,
	"ILL":      syscall.SIGILL,
	"INT":      syscall.SIGINT,
	"IO":       syscall.SIGIO,
	"IOT":      syscall.SIGIOT,
	"KILL":     syscall.SIGKILL,
	"PIPE":     syscall.SIGPIPE,
	"POLL":     syscall.SIGPOLL,
	"PROF":     syscall.SIGPROF,
	"PWR":      syscall.SIGPWR,
	"QUIT":     syscall.SIGQUIT,
	"SEGV":     syscall.SIGSEGV,
	"STKFLT":   syscall.SIGSTKFLT,
	"STOP":     syscall.SIGSTOP,
	"SYS":      syscall.SIGSYS,
	"TERM":     syscall.SIGTERM,
	"TRAP":     syscall.SIGTRAP,
	"TSTP":     syscall.SIGTSTP,
	"TTIN":     syscall.SIGTTIN,
	"TTOU":     syscall.SIGTTOU,
	"UNUSED":   syscall.SIGUNUSED,
	"URG":      syscall.SIGURG,
	"USR1":     syscall.SIGUSR1,
	"USR2":     syscall.SIGUSR2,
	"VTALRM":   syscall.SIGVTALRM,
	"WINCH":    syscall.SIGWINCH,
	"XCPU":     syscall.SIGXCPU,
	"XFSZ":     syscall.SIGXFSZ,
	"RTMIN":    sigrtmin,
	"RTMIN+1":  sigrtmin + 1,
	"RTMIN+2":  sigrtmin + 2,
	"RTMIN+3":  sigrtmin + 3,
	"RTMIN+4":  sigrtmin + 4,
	"RTMIN+5":  sigrtmin + 5,
	"RTMIN+6":  sigrtmin + 6,
	"RTMIN+7":  sigrtmin + 7,
	"RTMIN+8":  sigrtmin + 8,
	"RTMIN+9":  sigrtmin + 9,
	"RTMIN+10": sigrtmin + 10,
	"RTMIN+11": sigrtmin + 11,
	"RTMIN+12": sigrtmin + 12,
	"RTMIN+13": sigrtmin + 13,
	"RTMIN+14": sigrtmin + 14,
	"RTMIN+15": sigrtmin + 15,
	"RTMAX-14": sigrtmax - 14,
	"RTMAX-13": sigrtmax - 13,
	"RTMAX-12": sigrtmax - 12,
	"RTMAX-11": sigrtmax - 11,
	"RTMAX-10": sigrtmax - 10,
	"RTMAX-9":  sigrtmax - 9,
	"RTMAX-8":  sigrtmax - 8,
	"RTMAX-7":  sigrtmax - 7,
	"RTMAX-6":  sigrtmax - 6,
	"RTMAX-5":  sigrtmax - 5,
	"RTMAX-4":  sigrtmax - 4,
	"RTMAX-3":  sigrtmax - 3,
	"RTMAX-2":  sigrtmax - 2,
	"RTMAX-1":  sigrtmax - 1,
	"RTMAX":    sigrtmax,
}

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
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
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

// StringSliceEqual compares two string slice, ignore the order.
// If all items in the two string slice are equal, this function will return true
// even though there may have duplicate elements in the slice, otherwise reture false.
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

	// mapKeys to remember keys that exist in s1
	mapKeys := map[string]int{}

	// first list all items in s1
	for _, v := range s1 {
		mapKeys[v]++
	}

	// second list all items in s2
	for _, v := range s2 {
		mapKeys[v]--

		// we may get -1 in two cases:
		// 1. the item exists in the s2, but not in the s1;
		// 2. the item exists both in s1 and s2, but has different copies.
		// Under the condition that the length of slices are equals,
		// so we can quickly return false.
		if mapKeys[v] < 0 {
			return false
		}
	}

	return true
}

// MergeMap merges the m2 into m1, if it has the same keys, m2 will overwrite m1.
func MergeMap(m1 map[string]interface{}, m2 map[string]interface{}) (map[string]interface{}, error) {
	if m1 == nil && m2 == nil {
		return nil, fmt.Errorf("all of maps are nil")
	}

	if m1 == nil {
		return m2, nil
	}

	if m2 == nil {
		return m1, nil
	}

	for k, v := range m2 {
		m1[k] = v
	}

	return m1, nil
}

// StringDefault return default value if s is empty, otherwise return s.
func StringDefault(s string, val string) string {
	if s != "" {
		return s
	}
	return val
}

// ToStringMap changes the map[string]interface{} to map[string]string,
// If the interface is not string, it will be ignore.
func ToStringMap(in map[string]interface{}) map[string]string {
	if in == nil {
		return nil
	}
	out := make(map[string]string)
	for k, v := range in {
		if s, ok := v.(string); ok {
			out[k] = s
		}
	}
	return out
}

// StringSliceDelete deletes the `del` string in string slice.
func StringSliceDelete(in []string, del string) []string {
	if in == nil {
		return nil
	}

	out := make([]string, 0)
	for _, value := range in {
		if value != del {
			out = append(out, value)
		}
	}

	return out
}

// ResolveHomeDir resolve a target path from home dir, home dir must not be a relative
// path, must not be a file, create directory if not exist, returns the target
// directory if directory is symlink.
func ResolveHomeDir(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("home dir should not be empty")
	}
	if !filepath.IsAbs(path) {
		return "", fmt.Errorf("home dir %s should be an absolute path", path)
	}

	// create directory for home-dir if is not exist, or check if exist home-dir
	// is directory.
	if pinfo, err := os.Stat(path); err != nil {
		if err := os.MkdirAll(path, 0666); err != nil {
			return "", fmt.Errorf("failed to mkdir for home dir %s: %v", path, err)
		}
	} else if !pinfo.Mode().IsDir() {
		return "", fmt.Errorf("home dir %s should be directory", path)
	}

	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("failed to acquire real path for %s: %s", path, err)
	}

	return realPath, nil
}

// MatchLabelSelector returns true if labels cover selector.
func MatchLabelSelector(selector, labels map[string]string) bool {
	for k, v := range selector {
		if val, ok := labels[k]; ok {
			if v != val {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

// ExtractIPAndPortFromAddresses extract first valid ip and port from addresses.
func ExtractIPAndPortFromAddresses(addresses []string) (string, string) {
	for _, addr := range addresses {
		addrParts := strings.SplitN(addr, "://", 2)
		if len(addrParts) != 2 {
			logrus.Errorf("invalid listening address %s: must be in format [protocol]://[address]", addr)
			continue
		}

		switch addrParts[0] {
		case "tcp":
			host, port, err := net.SplitHostPort(addrParts[1])
			if err != nil {
				logrus.Errorf("failed to split host and port from address: %v", err)
				continue
			}
			return host, port
		case "unix":
			continue
		default:
			logrus.Errorf("only unix socket or tcp address is support")
		}
	}
	return "", ""
}

// ParseSignal translates a string to a valid syscall signal.
// It returns an error if the signal map doesn't include the given signal.
func ParseSignal(rawSignal string) (syscall.Signal, error) {
	s, err := strconv.Atoi(rawSignal)
	if err == nil {
		if s == 0 {
			return -1, fmt.Errorf("Invalid signal: %s", rawSignal)
		}
		return syscall.Signal(s), nil
	}
	signal, ok := SignalMap[strings.TrimPrefix(strings.ToUpper(rawSignal), "SIG")]
	if !ok {
		return -1, fmt.Errorf("Invalid signal: %s", rawSignal)
	}
	return signal, nil
}
