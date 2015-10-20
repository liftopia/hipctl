package main

import (
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
	"github.com/op/go-logging"
	"github.com/soveran/redisurl"
)

var conn redis.Conn

var log = logging.MustGetLogger("hipctl")
var format = logging.MustStringFormatter(
	"%{color}%{time:20060102 15:04:05.000} %{shortfunc:-20s} ▶ %{level:-6s} %{id:03x}%{color:reset} %{message}",
)

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

func init() {
	logbackend := logging.NewLogBackend(os.Stderr, "", 0)
	logbackendformatter := logging.NewBackendFormatter(logbackend, format)
	logging.SetBackend(logbackendformatter)
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
							log.Debug("adding %s to %v", ip, fe.key)
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

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	log.Notice("%v %v %v %v", mem.Alloc, mem.TotalAlloc, mem.HeapAlloc, mem.HeapSys)
	err := app.Run(os.Args)
	if err != nil {
		log.Error("%+v", err)
	}
}
