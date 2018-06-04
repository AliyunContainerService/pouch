package bytefmt

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	// BYTE represents one byte.
	BYTE = 1.0

	// KILOBYTE represents kilo bytes.
	KILOBYTE = 1024

	// MEGABYTE represents mega bytes.
	MEGABYTE = 1024 * KILOBYTE

	//GIGABYTE represents giga bytes.
	GIGABYTE = 1024 * MEGABYTE

	// TERABYTE represents tera bytes.
	TERABYTE = 1024 * GIGABYTE
)

var bytesPattern = regexp.MustCompile(`(?i)^(-?\d+(?:\.\d+)?)([KMGT]B?|B)$`)

// ErrorInvalidByte is the error that presents the format of string is not a valid byte.
var ErrorInvalidByte = errors.New("Byte quantity must be a positive integer with a unit of measurement like M, MB, G, or GB")

// ByteSize returns a human-readable byte string of the form 10M, 12.5K, and so forth.  The following units are available:
//	T: Terabyte
//	G: Gigabyte
//	M: Megabyte
//	K: Kilobyte
//	B: Byte
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
func ByteSize(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= TERABYTE:
		unit = "T"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimSuffix(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}

// ToMegabytes parses a string formatted by ByteSize as megabytes.
func ToMegabytes(s string) (uint64, error) {
	bytes, err := ToBytes(s)
	if err != nil {
		return 0, err
	}

	return bytes / MEGABYTE, nil
}

// ToKilobytes parses a string formatted by ByteSize as kilobytes.
func ToKilobytes(s string) (uint64, error) {
	bytes, err := ToBytes(s)
	if err != nil {
		return 0, err
	}

	return bytes / KILOBYTE, nil
}

// ToBytes parses a string formatted by ByteSize as bytes.
func ToBytes(s string) (uint64, error) {
	l := len(s)
	if l < 1 {
		return 0, ErrorInvalidByte
	}

	if s[l-1] != 'b' && s[l-1] != 'B' {
		s = s + "B"
	}

	parts := bytesPattern.FindStringSubmatch(strings.TrimSpace(s))
	if len(parts) < 3 {
		return 0, ErrorInvalidByte
	}

	value, err := strconv.ParseFloat(parts[1], 64)
	if err != nil || value <= 0 {
		return 0, ErrorInvalidByte
	}

	var bytes uint64
	unit := strings.ToUpper(parts[2])
	switch unit[:1] {
	case "T":
		bytes = uint64(value * TERABYTE)
	case "G":
		bytes = uint64(value * GIGABYTE)
	case "M":
		bytes = uint64(value * MEGABYTE)
	case "K":
		bytes = uint64(value * KILOBYTE)
	case "B":
		bytes = uint64(value * BYTE)
	}

	return bytes, nil
}
