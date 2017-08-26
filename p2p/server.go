package p2p

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"sync"

	"github.com/grandcat/zeroconf"
)

type Server struct {
	Port         int
	InstanceName string
	ServiceTag   string
	Duration     int
	src          string
	connNum      int
	errCh        chan *clientErr
	finishedCh   chan net.Addr

	mu          sync.Mutex
	clientConns map[net.Addr]*net.TCPConn
}

type clientErr struct {
	addr net.Addr
	err  error
}

func NewServer(instance, tag, file string, duration, num int) *Server {
	server := &Server{
		Port:         freePort(),
		InstanceName: instance,
		ServiceTag:   tag,
		Duration:     duration,
		src:          file,
		connNum:      num,
		errCh:        make(chan *clientErr),
		finishedCh:   make(chan net.Addr),
		clientConns:  make(map[net.Addr]*net.TCPConn),
	}
	return server
}

func (s *Server) Open() error {
	addr, err := net.ResolveTCPAddr("tcp", ":"+strconv.Itoa(s.Port))
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
	server, err := zeroconf.Register(s.InstanceName, s.ServiceTag, "local.", s.Port, []string{"txtv=0", "lo=1", "la=2"}, nil)
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
		s.errCh <- &clientErr{conn.RemoteAddr(), err}
	}
	defer f.Close()

	buf := make([]byte, 1024)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			s.finishedCh <- conn.RemoteAddr()
			break
		}
		if err != nil {
			s.errCh <- &clientErr{conn.RemoteAddr(), err}
			return
		}
		conn.Write(buf[:n])
	}
}
