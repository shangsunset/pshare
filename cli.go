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
	file := flag.String("file", "", "path to source file")
	flag.Parse()

	serviceTag := "_foobar._tcp"
	waitTime := 5
	clientNum := 0

	if *file != "" {

		s := p2p.NewServer(serviceTag, *file, waitTime, clientNum)

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

		go s.Register()

	} else {
		fmt.Println("connecting...")
		p2p.Connect(serviceTag, waitTime)
	}
}
