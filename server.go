package main

import (
	"fmt"
	"net"
	"net/url"

	"github.com/codegangsta/cli"
)

// Server stores the underlying server routing information
type Server struct {
	IP       net.IP
	Backends map[url.URL]*Backend
}

var servers []*Server

func (s *Server) String() string {
	return s.IP.String()
}

// ListServers prints out the list of servers
func ListServers(c *cli.Context) {
	for _, server := range servers {
		fmt.Printf("%+v\n", server)
	}
}

// AddBackend appends a known backend to the server's list
func (s *Server) AddBackend(b *Backend) {
	s.Backends[*b.Endpoint] = b
}

// ListServersComplete prints the server list for shell completions
func ListServersComplete(c *cli.Context) {
	if len(c.Args()) > 0 {
		return
	}
	for _, server := range servers {
		fmt.Println(server.IP.String())
	}
}

// NewServer grabs a specific server by IP
func NewServer(ip net.IP) (s *Server) {
	found := false

	for _, server := range servers {
		if server.IP.Equal(ip) {
			s = server
			found = true
		}
	}

	if !found {
		s = &Server{IP: ip}
		s.Backends = make(map[url.URL]*Backend)
		servers = append(servers, s)
	}

	return
}
