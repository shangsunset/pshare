package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shangsunset/pshare"
	"github.com/shangsunset/pshare/utils"
	"github.com/urfave/cli"
)

func main() {

	const waitTime = 500

	// Clean exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	// Timeout timer.
	var tc <-chan time.Time
	if waitTime > 0 {
		tc = time.After(time.Second * time.Duration(waitTime))
	}

	go func() {
		select {
		case <-sig:
			// Exit by user
			fmt.Println("Exited")
			os.Exit(1)
		case <-tc:
			// Exit by timeout
			fmt.Println("Timed out")
			os.Exit(1)
		}
	}()

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
				var instance, service, src string
				var clientNum int

				service = utils.RandString(20)
				instance = utils.RandString(20)
				src = c.Args().First()

				fmt.Println("Your service hash to share:", service)
				if c.Bool("private") {
					clientNum = 1
					fmt.Println("Your private instance hash to share:", instance)
				}

				s := pshare.NewServer(instance, service, src, waitTime, clientNum)
				err := s.Register()
				if err != nil {
					return cli.NewExitError(fmt.Errorf("Cli: %v", err), 0)
				}
				return nil
			},
		},
		{
			Name:    "recv",
			Aliases: []string{"r"},
			Usage:   "receive content from peer",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "service, s",
					Usage: "service hash",
				},
				cli.StringFlag{
					Name:  "private, p",
					Usage: "hash for private instance",
				},
			},
			Action: func(c *cli.Context) error {
				instance := c.String("instance")
				service := c.String("service")
				client := pshare.NewClient(instance, service, waitTime)
				if instance != "" {
					client.Lookup()
				} else {
					client.Browse()
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
