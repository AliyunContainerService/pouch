package jsonfile

import (
	"bytes"
	"encoding/json"
	"io"
	"time"
	"unicode/utf8"

	"github.com/alibaba/pouch/daemon/logger"
	"github.com/alibaba/pouch/pkg/utils"
)

type jsonLog struct {
	Source    string    `json:"source,omitempty"`
	Line      string    `json:"line,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

func newUnmarshal(r io.Reader) func() (*logger.LogMessage, error) {
	dec := json.NewDecoder(r)

	return func() (*logger.LogMessage, error) {
		jl := &jsonLog{}
		if err := dec.Decode(jl); err != nil {
			return nil, err
		}

		return &logger.LogMessage{
			Source:    jl.Source,
			Line:      []byte(jl.Line),
			Timestamp: jl.Timestamp,
		}, nil
	}
}

func marshal(msg *logger.LogMessage) ([]byte, error) {
	var (
		first = true
		buf   bytes.Buffer
	)

	buf.WriteString("{")
	if len(msg.Source) != 0 {
		first = false
		buf.WriteString(`"source":`)
		bytesIntoJSONString(&buf, []byte(msg.Source))
	}

	if len(msg.Line) != 0 {
		if !first {
			buf.WriteString(`,`)
		}
		first = false
		buf.WriteString(`"line":`)
		bytesIntoJSONString(&buf, msg.Line)
	}

	if !first {
		buf.WriteString(`,`)
	}

	buf.WriteString(`"timestamp":`)
	buf.WriteString(msg.Timestamp.UTC().Format(`"` + utils.TimeLayout + `"`))
	buf.WriteString(`}`)

	// NOTE: add newline here to make the decoder easier
	buf.WriteByte('\n')

	bs := buf.Bytes()
	buf.Reset()
	return bs, nil
}

// bytesIntoJSONString copies from encoding/json/encode.go#stringBytes
func bytesIntoJSONString(buf *bytes.Buffer, bs []byte) {
	var hex = "0123456789abcdef"

	buf.WriteByte('"')
	start := 0
	for i := 0; i < len(bs); {
		if b := bs[i]; b < utf8.RuneSelf {
			// This encodes bytes < 0x20 except for \t, \n and \r.
			// If escapeHTML is set, it also escapes <, >, and &
			// because they can lead to security holes when
			// user-controlled strings are rendered into JSON
			// and served to some browsers.
			if 0x20 <= b && b != '\\' && b != '"' && b != '<' && b != '>' && b != '&' {
				i++
				continue

			}

			if start < i {
				buf.Write(bs[start:i])
			}
			switch b {
			case '\\', '"':
				buf.WriteByte('\\')
				buf.WriteByte(b)
			case '\n':
				buf.WriteByte('\\')
				buf.WriteByte('n')
			case '\r':
				buf.WriteByte('\\')
				buf.WriteByte('r')
			case '\t':
				buf.WriteByte('\\')
				buf.WriteByte('t')
			default:
				buf.WriteString(`\u00`)
				buf.WriteByte(hex[b>>4])
				buf.WriteByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}

		c, size := utf8.DecodeRune(bs[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				buf.Write(bs[start:i])
			}
			buf.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}

		// U+2028 is LINE SEPARATOR.
		// U+2029 is PARAGRAPH SEPARATOR.
		// They are both technically valid characters in JSON strings,
		// but don't work in JSONP, which has to be evaluated as JavaScript,
		// and can lead to security holes there. It is valid JSON to
		// escape them, so we do so unconditionally.
		// See http://timelessrepo.com/json-isnt-a-javascript-subset for discussion.
		if c == '\u2028' || c == '\u2029' {
			if start < i {
				buf.Write(bs[start:i])
			}
			buf.WriteString(`\u202`)
			buf.WriteByte(hex[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(bs) {
		buf.Write(bs[start:])
	}
	buf.WriteByte('"')
}
