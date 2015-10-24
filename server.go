package main

import (
	"fmt"
	"net"

	"github.com/codegangsta/cli"
)

// Server stores the underlying server routing information
type Server struct {
	IP       net.IP
	Backends backendpool
}

var servers []*Server

func (s *Server) String() string {
	return fmt.Sprintf("%s, %d backends", s.IP, len(s.Backends))
}

func listServers() {
	for _, server := range servers {
		fmt.Println(server)
	}
	fmt.Printf("%d servers\n", len(servers))
}

// ShowServer gives detailed information about a server
func ShowServer(host string) {
	if s := GetServer(net.ParseIP(host)); s == nil {
		fmt.Printf("Couldn't find server %s\n", host)
	} else {
		fmt.Printf("%s\n%s\n", s, s.Backends.list())
	}
}

// AddBackend appends a known backend to the server's list
func (s *Server) AddBackend(b *Backend) {
	s.Backends = append(s.Backends, b)
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
	if s = GetServer(ip); s != nil {
		log.Debugf("Found with GetServer: %v", s)
		return
	}

	log.Debugf("Passing off to CreateServer: %s", ip)
	return CreateServer(ip)
}

// CreateServer handles the actual generation of a new server from IP
func CreateServer(ip net.IP) (s *Server) {
	log.Debugf("Creating new server at %s", ip)
	s = &Server{IP: ip}
	servers = append(servers, s)
	log.Debugf("Server added! %v, (%d) %#v", s, len(servers), servers)

	return
}

// GetServer grabs a specific server by IP without creating a new one
func GetServer(ip net.IP) (s *Server) {
	log.Debugf("Getting server at %s", ip)
	for _, server := range servers {
		if server.IP.Equal(ip) {
			s = server
		}
	}

	return
}
