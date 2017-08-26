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

func connect(host string, port int) {
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

func Connect(serviceTag string, waitTime int) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}
	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			fmt.Println(entry)
			connect(entry.HostName, entry.Port)
		}
		log.Println("No more entries.")
	}(entries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(waitTime))
	defer cancel()
	err = resolver.Browse(ctx, serviceTag, "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
	// Wait some additional time to see debug messages on go routine shutdown.
	// time.Sleep(1 * time.Second)
}
