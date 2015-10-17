package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/soveran/redisurl"
)

var conn redis.Conn
var frontends map[string]frontend
var backends map[string]backend

func validateips(c *cli.Context) error {
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

	return nil
}

// yeah i know
func setupglobals(c *cli.Context) (err error) {
	conn, err = redisurl.ConnectToURL(c.GlobalString("redis"))
	if err != nil {
		return
	}

	frontends, err = getfrontends()
	if err != nil {
		return
	}
	if len(frontends) == 0 {
		return errors.New("empty frontends list :(")
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "hipctl"
	app.Usage = "hipache bulk manager"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "redis",
			Value:  "redis://127.0.0.1/6379",
			EnvVar: "REDIS_URL",
		},
	}

	app.Before = setupglobals
	app.Commands = []cli.Command{
		{
			Name:  "list",
			Usage: "list frontends and backends",
			Action: func(c *cli.Context) {
				for _, fe := range frontends {
					fmt.Printf("%v\n", &fe)
					fmt.Println(strings.Repeat("-", 120))
				}
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
						if fe.hasbackend(ip) {
							fe.removebackend(ip)
						}
					}
				}
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
