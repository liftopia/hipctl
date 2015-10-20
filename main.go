package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/soveran/redisurl"
)

var conn redis.Conn

var log = logrus.New()
var formatter = &logrus.TextFormatter{
	FullTimestamp: true,
}

func validateips(c *cli.Context) (err error) {
	if len(c.Args()) == 0 {
		return errors.New("IP(s) required")
	}

	var badips []string
	for _, ip := range c.Args() {
		if net.ParseIP(ip) == nil {
			badips = append(badips, ip)
		}
	}

	if len(badips) > 0 {
		return fmt.Errorf("Bad IPs: %v", strings.Join(badips, ", "))
	}

	return
}

// yeah i know
func setupglobals(c *cli.Context) (err error) {
	if !c.GlobalBool("debug") {
		log.Level = logrus.ErrorLevel
	}
	conn, err = redisurl.ConnectToURL(c.GlobalString("redis"))
	if err != nil {
		return
	}

	updatefrontends()
	if len(frontends) == 0 {
		return errors.New("empty frontends list :(")
	}

	return
}

func setuplogs() {
	log.Formatter = formatter
	log.Out = os.Stderr
	log.Level = logrus.DebugLevel
}

func init() {
	setuplogs()
}

func main() {
	app := cli.NewApp()
	app.EnableBashCompletion = true
	app.Name = "hipctl"
	app.Usage = "hipache bulk manager"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "redis",
			Value:  "redis://127.0.0.1/6379",
			EnvVar: "REDIS_URL",
		},
		cli.BoolFlag{
			Name:   "debug",
			EnvVar: "CONTROL_DEBUG",
		},
	}

	app.Before = setupglobals
	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "list frontends, backends, and more",
			Action: func(c *cli.Context) {
				for _, fe := range frontends {
					fmt.Printf("%v\n", &fe)
					for _, be := range fe.Backends {
						fmt.Printf("%v\n", be)
					}
					fmt.Println(strings.Repeat("-", 120))
				}
			},
			Subcommands: []cli.Command{
				{
					Name:         "servers",
					Usage:        "list configured servers",
					BashComplete: ListServersComplete,
					Action:       ListServers,
				},
			},
		},
		{
			Name:  "show",
			Usage: "show a particular system",
			Subcommands: []cli.Command{
				{
					Name:         "frontend",
					Usage:        "show frontend information",
					BashComplete: ListFrontendsComplete,
					Before: func(c *cli.Context) error {
						if len(c.Args()) != 1 {
							return errors.New("Incorrect parameters.")
						}
						return nil
					},
					Action: func(c *cli.Context) {
						ShowFrontend(c.Args().First())
					},
				},
				{
					Name:         "server",
					Usage:        "show server information",
					BashComplete: ListServersComplete,
					Before: func(c *cli.Context) error {
						if len(c.Args()) != 1 {
							return errors.New("Incorrect parameters.")
						}
						return nil
					},
					Action: func(c *cli.Context) {
						ShowServer(c.Args().First())
					},
				},
			},
		},
		{
			Name:   "add",
			Usage:  "add backend by ip",
			Before: validateips,
			Action: func(c *cli.Context) {
				for _, ip := range c.Args() {
					for _, fe := range frontends {
						if !fe.hasbackend(ip) {
							fe.addbackend(ip)
						}
					}
				}
			},
		},
		{
			Name:   "remove",
			Usage:  "remove backend by ip",
			Before: validateips,
			Action: func(c *cli.Context) {
				for _, ip := range c.Args() {
					for _, fe := range frontends {
						if be := fe.getbackend(ip); be != nil {
							fe.removebackend(be)
						}
					}
				}
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("%+v", err)
	}
}
