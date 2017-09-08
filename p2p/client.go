package p2p

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
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
			fmt.Println(entry)
			c.connect(entry.HostName, entry.Port)
		}
		log.Println("No more entries.")
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
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
}

func (c *Client) connect(host string, port int) {
	addr, err := net.ResolveTCPAddr("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		_, err = conn.Read(buffer)
		if err == io.EOF {
			fmt.Println("Done!")
			break
		} else if err != nil {
			log.Fatal(err)
		}
		// fmt.Println(string(buffer))
	}
}
