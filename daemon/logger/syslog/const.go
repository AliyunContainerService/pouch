package syslog

import (
	"crypto/tls"

	"github.com/RackSec/srslog"
)

var (
	// rfc5424 provides millisecond resolution.
	timeRfc5424fmt = "2006-01-02T15:04:05.999999Z07:00"

	secureProto = "tcp+tls"

	defaultTagTemplate    = "{{.ID}}"
	defaultSyslogPriority = srslog.LOG_DAEMON

	// facilityAliasMap allows user to use alias to set the syslog priority.
	facilityAliasMap = map[string]srslog.Priority{
		"kern":     srslog.LOG_KERN,
		"user":     srslog.LOG_USER,
		"mail":     srslog.LOG_MAIL,
		"daemon":   srslog.LOG_DAEMON,
		"auth":     srslog.LOG_AUTH,
		"syslog":   srslog.LOG_SYSLOG,
		"lpr":      srslog.LOG_LPR,
		"news":     srslog.LOG_NEWS,
		"uucp":     srslog.LOG_UUCP,
		"cron":     srslog.LOG_CRON,
		"authpriv": srslog.LOG_AUTHPRIV,
		"ftp":      srslog.LOG_FTP,
		"local0":   srslog.LOG_LOCAL0,
		"local1":   srslog.LOG_LOCAL1,
		"local2":   srslog.LOG_LOCAL2,
		"local3":   srslog.LOG_LOCAL3,
		"local4":   srslog.LOG_LOCAL4,
		"local5":   srslog.LOG_LOCAL5,
		"local6":   srslog.LOG_LOCAL6,
		"local7":   srslog.LOG_LOCAL7,
	}

	validTransportURLPrefix = []string{
		"udp://",
		"tcp://",
		"tcp+tls://",
		"unix://",
		"unixgram://",
	}

	// tls client cipher suites
	defaultCipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}
)
