package utils

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Common durations that is .
// There are some definitions for units of Day and larger .
const (
	Second     = time.Second
	Minute     = Second * 60
	Hour       = Minute * 60
	Day        = Hour * 24
	Week       = Day * 7
	Month      = Day * 30
	Year       = Day * 365
	TimeLayout = "2006-01-02 15:04:05"
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
func FormatTimeInterval(input string) (formattedTime string, err error) {
	start, err := time.Parse(TimeLayout, input)
	if err != nil {
		return "", errInvalid
	}
	diff := time.Now().Sub(start)

	// That should not happen.
	if diff < 0 {
		formattedTime += "-"
		diff = 0 - diff
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
