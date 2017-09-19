package pshare

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/grandcat/zeroconf"
	"github.com/shangsunset/pshare/utils"
)

type Server struct {
	port       int
	instance   string
	service    string
	duration   int
	src        string
	connNum    int
	errCh      chan *connErr
	finishedCh chan net.Addr

	mu          sync.Mutex
	clientConns map[net.Addr]*net.TCPConn
}

type connErr struct {
	addr net.Addr
	err  error
}

func NewServer(instance, service, src string, duration, num int) *Server {
	server := &Server{
		port:        utils.RandPort(),
		instance:    instance,
		service:     service,
		duration:    duration,
		src:         src,
		connNum:     num,
		errCh:       make(chan *connErr),
		finishedCh:  make(chan net.Addr),
		clientConns: make(map[net.Addr]*net.TCPConn),
	}
	return server
}

func (s *Server) Register() error {
	server, err := zeroconf.Register(s.instance, s.service, "local.", s.port, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		return fmt.Errorf("Zeroconf: %v", err)
	}
	defer server.Shutdown()

	err = s.Open()
	if err != nil {
		return fmt.Errorf("Failed to open connection: %v\n", err)
	}

	return nil
}

func (s *Server) Open() error {
	addr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(s.port))
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

func (s *Server) Serve(ln *net.TCPListener) error {
	finished := 0
	go func() {
		for {
			select {
			case cerr := <-s.errCh:
				fmt.Errorf("Error from %v: %v\n", cerr.addr, cerr.err)
				err := s.clientConns[cerr.addr].Close()
				if err != nil {
					fmt.Errorf("Failded to close client connection: %v\n", err)
				}
				return
			case c := <-s.finishedCh:
				fmt.Printf("%v is done receiving\n", c)
				err := s.clientConns[c].Close()
				if err != nil {
					fmt.Errorf("Failded to close client connection: %v\n", err)
				}
				s.mu.Lock()
				s.clientConns[c].Close()
				finished++
				s.mu.Unlock()
				if finished == s.connNum {
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
		conn, err := ln.AcceptTCP()
		if err != nil {
			return err
		}
		s.clientConns[conn.RemoteAddr()] = conn
		go s.handleConn(conn)
	}
	return nil
}

func (s *Server) handleConn(conn *net.TCPConn) {
	f, err := os.Open(s.src)
	if err != nil {
		s.errCh <- &connErr{conn.RemoteAddr(), err}
	}
	defer f.Close()

	sep := ";;;"
	_, err = conn.Write([]byte(s.src + sep))
	if err != nil {
		s.errCh <- &connErr{conn.RemoteAddr(), fmt.Errorf("Initial write failed: %v", err)}
	}
	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			s.finishedCh <- conn.RemoteAddr()
			break
		}
		if err != nil {
			s.errCh <- &connErr{conn.RemoteAddr(), err}
			return
		}
		_, err = conn.Write(buf[:n])
		if err != nil {
			s.errCh <- &connErr{conn.RemoteAddr(), err}
		}
	}
}
