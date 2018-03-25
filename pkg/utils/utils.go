package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
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

	if diff >= Year {
		year := int(diff / Year)
		formattedTime += strconv.Itoa(year) + " year"
		if year > 1 {
			formattedTime += "s"
		}
	} else if diff >= Month {
		month := int(diff / Month)
		formattedTime += strconv.Itoa(month) + " month"
		if month > 1 {
			formattedTime += "s"
		}
	} else if diff >= Week {
		week := int(diff / Week)
		formattedTime += strconv.Itoa(week) + " week"
		if week > 1 {
			formattedTime += "s"
		}
	} else if diff >= Day {
		day := int(diff / Day)
		formattedTime += strconv.Itoa(day) + " day"
		if day > 1 {
			formattedTime += "s"
		}
	} else if diff >= Hour {
		hour := int(diff / Hour)
		formattedTime += strconv.Itoa(hour) + " hour"
		if hour > 1 {
			formattedTime += "s"
		}
	} else if diff >= Minute {
		minute := int(diff / Minute)
		formattedTime += strconv.Itoa(minute) + " minute"
		if minute > 1 {
			formattedTime += "s"
		}
	} else if diff >= Second {
		second := int(diff / Second)
		formattedTime += strconv.Itoa(second) + " second"
		if second > 1 {
			formattedTime += "s"
		}
	} else {
		formattedTime += "0 second"
	}

	return formattedTime, nil
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

// doMerge, begin merge action
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
			if err := doMerge(src.MapIndex(key), dest.MapIndex(key)); err != nil {
				return err
			}
		}

	default:
		dest.Set(src)
	}

	return nil
}

// From src/pkg/encoding/json
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
