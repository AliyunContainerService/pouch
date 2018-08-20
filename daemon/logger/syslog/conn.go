package syslog

import (
	"net"

	"github.com/RackSec/srslog"
)

// localConn implements the serverConn interface, used to send syslog messages
// to the remote syslog daemon.
type remoteConn struct {
	conn net.Conn
}

// writeString will use Framer/Formatter to format the content before write.
//
// NOTE: writeString implements serverConn.writeString() methods.
func (n *remoteConn) writeString(framer Framer, formatter Formatter, p Priority, hostname, tag, msg string) error {
	if framer == nil {
		framer = srslog.DefaultFramer
	}

	if formatter == nil {
		formatter = srslog.DefaultFormatter
	}

	formattedMessage := framer(formatter(p, hostname, tag, msg))
	_, err := n.conn.Write([]byte(formattedMessage))
	return err
}

// close closes the connection.
//
// NOTE:close implements serverConn.close() methods.
func (n *remoteConn) close() error {
	return n.conn.Close()
}

// localConn implements the serverConn interface, used to send syslog messages
// to the local syslog daemon over a Unix domain socket.
type localConn struct {
	conn net.Conn
}

// writeString will use Framer/Formatter to format the content before write.
//
// NOTE: writeString implements serverConn.writeString() methods.
func (n *localConn) writeString(framer Framer, formatter Formatter, p Priority, hostname, tag, msg string) error {
	if framer == nil {
		framer = srslog.DefaultFramer
	}

	if formatter == nil {
		formatter = srslog.UnixFormatter
	}

	formattedMessage := framer(formatter(p, hostname, tag, msg))
	_, err := n.conn.Write([]byte(formattedMessage))
	return err
}

// close closes the connection.
//
// NOTE:close implements serverConn.close() methods.
func (n *localConn) close() error {
	return n.conn.Close()
}
