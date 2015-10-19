package main

import (
	"net"
	"net/url"
)

// Server stores the underlying server routing information
type Server struct {
	IP       net.IP
	Backends map[url.URL]*Backend
}

var servers []*Server

// AddBackend appens a known backend to the server's list
func (s *Server) AddBackend(b *Backend) {
	s.Backends[*b.Endpoint] = b
}

// GetServer grabs a specific server by IP
func GetServer(ip string) (s *Server) {
	s = &Server{IP: net.ParseIP(ip)}
	if len(servers) > 0 {
		for _, server := range servers {
			if server.IP.String() == ip {
				s = server
			}
		}
	}
	return
}
