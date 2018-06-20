package main

import (
	"flag"
	"log"
	"net"
	"strings"
	"time"

	"github.com/ishidawataru/sctp"
)

func serveClient(conn net.Conn) error {
	for {
		buf := make([]byte, 254)
		n, err := conn.Read(buf)
		if err != nil {
			return err
		}
		n, err = conn.Write(buf[:n])
		if err != nil {
			return err
		}
		log.Printf("write: %d", n)
	}
}

func main() {
	var server = flag.Bool("server", false, "")
	var ip = flag.String("ip", "0.0.0.0", "")
	var port = flag.Int("port", 0, "")
	flag.Parse()

	ips := []net.IP{}

	for _, i := range strings.Split(*ip, ",") {
		ips = append(ips, net.ParseIP(i))
	}

	addr := &sctp.SCTPAddr{
		IP:   ips,
		Port: *port,
	}

	if *server {
		ln, err := sctp.ListenSCTP("sctp", addr)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Printf("Listen on %s", ln.Addr())

		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Fatalf("failed to accept: %v", err)
			}
			wconn := sctp.NewSCTPSndRcvInfoWrappedConn(conn.(*sctp.SCTPConn))
			go serveClient(wconn)
		}

	} else {

		conn, err := sctp.DialSCTP("sctp", nil, addr)
		if err != nil {
			log.Fatalf("failed to dial: %v", err)
		}
		ppid := 0
		for {
			info := &sctp.SndRcvInfo{
				Stream: uint16(ppid),
				PPID:   uint32(ppid),
			}
			ppid += 1
			conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)
			n, err := conn.SCTPWrite([]byte("hello"), info)
			if err != nil {
				log.Fatalf("failed to write: %v", err)
			}
			log.Printf("write: %d", n)
			buf := make([]byte, 254)
			_, info, err = conn.SCTPRead(buf)
			if err != nil {
				log.Fatalf("failed to read: %v", err)
			}
			log.Printf("read: info: %+v", info)
			time.Sleep(time.Second)
		}
	}
}
