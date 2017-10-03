package pshare

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/grandcat/zeroconf"
)

type Client struct {
	instanceName string
	serviceName  string
	waitTime     int
	resolver     *zeroconf.Resolver
	entries      chan *zeroconf.ServiceEntry
}

func NewClient(instance, service string, waitTime int) *Client {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		fmt.Println("failed to initialize resolver:", err.Error())
	}

	c := Client{
		instanceName: instance,
		serviceName:  service,
		waitTime:     waitTime,
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
	buf := make([]byte, 1024)

	_, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("Failed to receive file name: %v", err)
	}

	sep := ";;;"
	data := strings.Split(string(buf), sep)
	filename := data[0]

	var response string
	fmt.Fprintf(os.Stderr, "Accept %s? [Y/n] ", filename)
	fmt.Fscanf(os.Stderr, "%s", &response)

	if response != "y" && response != "Y" {
		return nil
	}

	fout, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fout.Close()

	w := bufio.NewWriter(fout)

	if _, err = w.Write([]byte(data[1])); err != nil {
		return fmt.Errorf("Initial write to buffer failed: %v", err)
	}
	if err = w.Flush(); err != nil {
		return err
	}

	for {
		n, err := conn.Read(buf)
		if err == io.EOF {
			fmt.Println("Done!")
			return nil
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
	return nil
}
