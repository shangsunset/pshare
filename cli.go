package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shangsunset/pshare/p2p"
)

func main() {
	src := flag.String("src", "", "path to source file")
	instanceName := flag.String("name", "", "service instance name")
	flag.Parse()

	serviceTag := "_foobar._tcp"
	waitTime := 50
	clientNum := 0
	if *src != "" {
		s := p2p.NewServer(*instanceName, serviceTag, *src, waitTime, clientNum)

		// Clean exit.
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		// Timeout timer.
		var tc <-chan time.Time
		if s.Duration > 0 {
			tc = time.After(time.Second * time.Duration(s.Duration))
		}

		go func() {
			select {
			case <-sig:
				// Exit by user
				log.Println("Exited")
				os.Exit(1)
			case <-tc:
				// Exit by timeout
				log.Println("Timed out")
				os.Exit(1)
			}
		}()

		err := s.Register()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println("connecting...")
		p2p.Connect(serviceTag, waitTime)
	}
}
