package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		{
			Name:    "share",
			Aliases: []string{"s"},
			Usage:   "share content with peers",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "private, p",
					Usage: "make the sharing private with only one peer",
				},
			},
			Action: func(c *cli.Context) error {
				fmt.Println("sharing is private?", c.Bool("private"))
				fmt.Println("sharing file:", c.Args())
				return nil
			},
		},
		{
			Name:    "recv",
			Aliases: []string{"r"},
			Usage:   "receive content from sender",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "instance, i",
					Usage: "service instance name",
				},
				cli.StringFlag{
					Name:  "service, s",
					Usage: "service name",
				},
			},
			Action: func(c *cli.Context) error {
				fmt.Println("instance", c.String("instance"))
				fmt.Println("service", c.String("service"))
				fmt.Println("receiving file:", c.Args())
				return nil
			},
		},
	}

	app.Run(os.Args)
	// serviceName := *serviceTag + "._tcp."
	// waitTime := 50
	// clientNum := 0
	//
	// // Clean exit.
	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	// // Timeout timer.
	// var tc <-chan time.Time
	// if waitTime > 0 {
	// 	tc = time.After(time.Second * time.Duration(waitTime))
	// }
	//
	// go func() {
	// 	select {
	// 	case <-sig:
	// 		// Exit by user
	// 		log.Println("Exited")
	// 		os.Exit(1)
	// 	case <-tc:
	// 		// Exit by timeout
	// 		log.Println("Timed out")
	// 		os.Exit(1)
	// 	}
	// }()
	//
	// if *src != "" {
	// 	s := p2p.NewServer(*instanceName, serviceName, *src, waitTime, clientNum)
	// 	err := s.Register()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// } else {
	// 	c := p2p.NewClient(*instanceName, serviceName, waitTime)
	// 	if instanceName != nil {
	// 		c.Lookup()
	// 	} else {
	// 		c.Browse()
	// 	}
	// }
}
