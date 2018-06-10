package syslog

import (
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"testing"

	"github.com/RackSec/srslog"
)

func TestParseFacility(t *testing.T) {
	for idx, tc := range []struct {
		input    string
		expected srslog.Priority
		err      error
	}{
		{
			input:    "",
			expected: defaultSyslogPriority,
			err:      nil,
		}, {
			input:    "local0",
			expected: srslog.LOG_LOCAL0,
			err:      nil,
		}, {
			input:    "1",
			expected: srslog.Priority(8),
			err:      nil,
		}, {
			input:    "invalid",
			expected: srslog.Priority(0),
			err:      ErrInvalidSyslogFacility,
		},
	} {
		got, err := parseFacility(tc.input)
		if err != tc.err {
			t.Fatalf("[%d case] expected error(%v), but got error(%v)", idx, tc.err, err)
		}

		if got != tc.expected {
			t.Fatalf("[%d case] expected priority(%v), but got(%v)", idx, tc.expected, got)
		}
	}
}

func TestParseTargetAddress(t *testing.T) {
	// create file for unix socket
	sockFile, err := ioutil.TempFile("", "testXXX.sock")
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		sockFile.Close()
		os.Remove(sockFile.Name())
	}()

	for idx, tc := range []struct {
		input          string
		proto, address string
		hasError       bool
	}{
		{
			input:    "http://localhost:8080",
			hasError: true,
		}, {
			input: "",
		}, {
			input:    "invalid url",
			hasError: true,
		}, {
			input:   "tcp://localhost:8080",
			proto:   "tcp",
			address: "localhost:8080",
		}, {
			input:   "udp://localhost",
			proto:   "udp",
			address: "localhost:514", // use default port
		}, {
			input:   "unix://" + sockFile.Name(),
			proto:   "unix",
			address: sockFile.Name(),
		}, {
			input:   "unixgram://" + sockFile.Name(),
			proto:   "unixgram",
			address: sockFile.Name(),
		}, {
			input:    "unixgram://" + sockFile.Name() + "dont_exist",
			hasError: true,
		},
	} {
		gotProto, gotAddress, err := parseTargetAddress(tc.input)
		if err != nil && !tc.hasError {
			t.Fatalf("[%d case] expect no error here, but got error: %v", idx, err)
		}

		if err == nil && tc.hasError {
			t.Fatalf("[%d case] expect error here, but got nothing", idx)
		}

		if gotProto != tc.proto {
			t.Fatalf("[%d case] expected proto(%v), but got(%v)", idx, tc.proto, gotProto)
		}

		if gotAddress != tc.address {
			t.Fatalf("[%d case] expected address(%v), but got(%v)", idx, tc.address, gotAddress)
		}
	}
}

func TestParseLogFormat(t *testing.T) {
	for idx, tc := range []struct {
		fmtTyp string
		proto  string

		formatter srslog.Formatter
		framer    srslog.Framer
		hasError  bool
	}{
		{
			fmtTyp:    "",
			formatter: srslog.UnixFormatter,
			framer:    srslog.DefaultFramer,
		}, {
			fmtTyp:    "rfc3164",
			proto:     secureProto,
			formatter: srslog.RFC3164Formatter,
			framer:    srslog.DefaultFramer,
		}, {
			fmtTyp:    "rfc5424",
			proto:     "tcp",
			formatter: rfc5424FormatterWithTagAsAppName,
			framer:    srslog.DefaultFramer,
		}, {
			fmtTyp:    "rfc5424",
			proto:     "tcp+tls",
			formatter: rfc5424FormatterWithTagAsAppName,
			framer:    srslog.RFC5425MessageLengthFramer,
		}, {
			fmtTyp:    "rfc5424micro",
			proto:     "tcp+tls",
			formatter: rfc5424MicroFormatterWithTagAsAppName,
			framer:    srslog.RFC5425MessageLengthFramer,
		}, {
			fmtTyp:    "rfc5424micro",
			proto:     "tcp",
			formatter: rfc5424MicroFormatterWithTagAsAppName,
			framer:    srslog.DefaultFramer,
		}, {
			fmtTyp:   "not support yet",
			hasError: true,
		},
	} {
		gotFmter, gotFramer, err := parseLogFormat(tc.fmtTyp, tc.proto)
		if err != nil && !tc.hasError {
			t.Fatalf("[%d case] expect no error here, but got error: %v", idx, err)
		}

		if err == nil && tc.hasError {
			t.Fatalf("[%d case] expect error here, but got nothing", idx)
		}

		if !isSameFunc(tc.formatter, gotFmter) || !isSameFunc(tc.framer, gotFramer) {
			t.Fatalf("[%d case] expect formatter(%v) & framer(%v), but got formatter(%v) & framer(%v)",
				idx, tc.formatter, tc.framer, gotFmter, gotFramer,
			)
		}
	}
}

func isSameFunc(aFunc interface{}, bFunc interface{}) bool {
	return reflect.ValueOf(aFunc).Pointer() == reflect.ValueOf(bFunc).Pointer()
}
