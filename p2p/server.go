package p2p

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	Port       int
	ServiceTag string
	FilePath   string
	Duration   int
	ClientNum  int
}

type Socket struct {
	conn net.Conn
	errs chan error
	done chan bool
}

func NewServer(tag, file string, duration, num int) *Server {
	server := &Server{
		Port:       9000,
		ServiceTag: tag,
		FilePath:   file,
		Duration:   duration,
		ClientNum:  num,
	}
	return server
}

func (s *Server) Open() error {
	addr, err := net.ResolveTCPAddr("tcp", ":9000")
	if nil != err {
		return err
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	fmt.Println("Listening on ", addr)

	err = s.Serve(ln)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Register() error {
	server, err := zeroconf.Register("my awesome service", s.ServiceTag, "local", s.Port, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		return err
	}
	defer server.Shutdown()

	err = s.Open()
	if err != nil {
		return fmt.Errorf("Failed to open connection: %v\n", err)
	}

	return nil
}

func (s *Server) Serve(ln *net.TCPListener) error {
	duration := time.Now().Add(time.Duration(s.Duration) * time.Second)
	// finished := 0
	for {
		ln.SetDeadline(duration)
		conn, err := ln.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				return fmt.Errorf("connection expired; lasted for %v", s.Duration)
			}
		}
		fmt.Println(conn.RemoteAddr())
		soc := &Socket{
			conn: conn,
			errs: make(chan error),
			done: make(chan bool),
		}
		go s.handleRequest(soc)
	}
	// select {
	// case err := <-soc.errs:
	// 	if err != nil {
	// 		return fmt.Errorf("error from streaming data: %v", err)
	// 	}
	// case <-soc.done:
	// 	log.Printf("%v is done\n", conn.RemoteAddr())
	// 	finished++
	// 	if finished == s.ClientNum {
	// 		soc.conn.Close()
	// 		return nil
	// 	}
	// }
	return nil
}

func (s *Server) handleRequest(soc *Socket) {
	f, err := os.Open(s.FilePath)
	if err != nil {
		soc.errs <- err
	}
	defer f.Close()

	buf := make([]byte, 6)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			soc.done <- true
			break
		}
		if err != nil {
			soc.errs <- err
			return
		}
		soc.conn.Write(buf[:n])
	}
}
