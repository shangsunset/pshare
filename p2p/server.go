package p2p

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	Port        int
	ServiceTag  string
	Source      string
	Duration    int
	ConnNum     int
	ClientConns map[net.Addr]*net.TCPConn
	ErrCh       chan *ClientErr
	FinishedCh  chan net.Addr

	mu sync.Mutex
}

type ClientErr struct {
	addr net.Addr
	err  error
}

func NewServer(tag, file string, duration, num int) *Server {
	server := &Server{
		Port:        9000,
		ServiceTag:  tag,
		Source:      file,
		Duration:    duration,
		ConnNum:     num,
		ErrCh:       make(chan *ClientErr),
		FinishedCh:  make(chan net.Addr),
		ClientConns: make(map[net.Addr]*net.TCPConn),
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
	// duration := time.Now().Add(time.Duration(s.Duration) * time.Second)
	finished := 0
	go func() {
		for {
			select {
			case cerr := <-s.ErrCh:
				fmt.Errorf("Error from %v: %v\n", cerr.addr, cerr.err)
				err := s.ClientConns[cerr.addr].Close()
				if err != nil {
					fmt.Errorf("Failded to close client connection: %v\n", err)
				}
				return
			case c := <-s.FinishedCh:
				fmt.Printf("%v is done receiving\n", c)
				err := s.ClientConns[c].Close()
				if err != nil {
					fmt.Errorf("Failded to close client connection: %v\n", err)
				}
				s.mu.Lock()
				s.ClientConns[c].Close()
				finished++
				s.mu.Unlock()
				if finished == s.ConnNum {
					err := ln.Close()
					if err != nil {
						fmt.Errorf("Failded to shut done listener: %v\n", err)
					}
					fmt.Println("connection closed")
				}
				return
			}
		}
	}()

	for {
		// ln.SetDeadline(duration)
		conn, err := ln.AcceptTCP()
		if err != nil {
			// if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
			// 	return fmt.Errorf("connection expired; lasted for %v", s.Duration)
			// }
			return err
		}
		s.ClientConns[conn.RemoteAddr()] = conn
		go s.handleConn(conn)
	}
	return nil
}

func (s *Server) handleConn(conn *net.TCPConn) {
	f, err := os.Open(s.Source)
	if err != nil {
		s.ErrCh <- &ClientErr{conn.RemoteAddr(), err}
	}
	defer f.Close()

	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			s.FinishedCh <- conn.RemoteAddr()
			break
		}
		if err != nil {
			s.ErrCh <- &ClientErr{conn.RemoteAddr(), err}
			return
		}
		conn.Write(buf[:n])
	}
}
