package p2p

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/grandcat/zeroconf"
)

type Client struct {
	instanceName string
	serviceName  string
	waitTime     int
	dest         string
	resolver     *zeroconf.Resolver
	entries      chan *zeroconf.ServiceEntry
}

func NewClient(instance, service string, dest string, waitTime int) *Client {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("failed to initialize resolver:", err.Error())
	}

	c := Client{
		instanceName: instance,
		serviceName:  service,
		waitTime:     waitTime,
		dest:         dest,
		resolver:     resolver,
		entries:      make(chan *zeroconf.ServiceEntry),
	}
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			err := c.connect(entry.HostName, entry.Port)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
			os.Exit(1)
		}
		fmt.Println("No more entries.")
	}(c.entries)
	return &c
}

func (c *Client) Lookup() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(c.waitTime))
	defer cancel()

	err := c.resolver.Lookup(ctx, c.instanceName, c.serviceName, "local.", c.entries)
	if err != nil {
		log.Fatalln("Failed to lookup:", err.Error())
	}

	<-ctx.Done()
}

func (c *Client) Browse() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(c.waitTime))
	defer cancel()

	err := c.resolver.Browse(ctx, c.serviceName, "local.", c.entries)
	if err != nil {
		log.Fatalln("Failed to browse entries:", err.Error())
	}

	<-ctx.Done()
}

func (c *Client) connect(host string, port int) error {
	addr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	err = c.receive(conn)
	if err != nil {
		return fmt.Errorf("Failed to receive data: %v", err)
	}
	return nil
}

func (c *Client) receive(conn *net.TCPConn) error {
	if _, err := os.Stat(c.dest); os.IsNotExist(err) {
		fout, err := os.Create(c.dest)
		if err != nil {
			return err
		}
		defer fout.Close()

		w := bufio.NewWriter(fout)
		buf := make([]byte, 1024)

		for {
			n, err := conn.Read(buf)
			if err == io.EOF {
				fmt.Println("Done!")
				break
			} else if err != nil {
				return err
			}
			if _, err := w.Write(buf[:n]); err != nil {
				return err
			}
			if err = w.Flush(); err != nil {
				return err
			}
		}
	} else {
		return errors.New("File already exists.")
	}
	return nil
}
