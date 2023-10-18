package rtmp

import (
	"bufio"
	"net"
	"time"
)

const (
	EventConnConnected     = 1
	EventHandshakeFailed   = 2
	EventConnDisconnected  = 4
	EventConnConnectFailed = 5
)

var EventString = map[int]string{
	EventConnConnected:     "Connected",
	EventConnConnectFailed: "ConnectFailed",
	EventHandshakeFailed:   "HandshakeFailed",
	EventConnDisconnected:  "ConnDisconnected",
}

var BufIoSize = 4096

type bufReadWriter struct {
	*bufio.Reader
	*bufio.Writer
}

type Server struct {
	HandshakeTimeout time.Duration // 握手超时时间

	LogEvent func(c *Conn, nc net.Conn, e int)

	HandleConn func(c *Conn, nc net.Conn)
}

func NewServer() *Server {
	return &Server{
		HandshakeTimeout: time.Second * 10,
	}
}

func (s *Server) HandleNetConn(nc net.Conn) {
	rw := &bufReadWriter{
		Reader: bufio.NewReaderSize(nc, BufIoSize),
		Writer: bufio.NewWriterSize(nc, BufIoSize),
	}
	c := NewConn(rw)

	if fn := s.LogEvent; fn != nil {
		fn(c, nc, EventConnConnected)
	}

	nc.SetDeadline(time.Now().Add(time.Second * 15))
	if err := c.Prepare(StageGotPublishOrPlayCommand); err != nil {
		if fn := s.LogEvent; fn != nil {
			fn(c, nc, EventHandshakeFailed)
		}
		nc.Close()
		return
	}
	nc.SetDeadline(time.Time{})

	s.HandleConn(c, nc)
}
