package main

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

// Backend stores the details about the frontend's connections
type Backend struct {
	Endpoint *url.URL
	Frontend *Frontend
	Server   *Server
}

// Port grabs the port from the Endpoint's Host key
func (b *Backend) Port() (port int) {
	port = 80
	if parts := strings.Split(b.Endpoint.Host, ":"); len(parts) > 1 {
		port, _ = strconv.Atoi(parts[1])
	}
	return
}

// Host grabs the host from the Endpoint's Host key
func (b *Backend) Host() (host string) {
	return strings.Split(b.Endpoint.Host, ":")[0]
}

func (b *Backend) String() string {
	return b.Endpoint.String()
}

// AddServer appends a known host to the backend's serving list
func (b *Backend) AddServer(s *Server) {
	b.Server = s
	s.AddBackend(b)
}

// Equal compares one Backend to another to see if they match
func (b *Backend) Equal(other *Backend) bool {
	if b.Endpoint == other.Endpoint && b.Frontend == other.Frontend {
		return true
	}
	return false
}

// Empty checks if the Backend is completely void of values
func (b *Backend) Empty() bool {
	if b.Endpoint == nil && b.Frontend == nil {
		return true
	}
	return false
}

// IsIP checks the Backend for the presence of the IP
func (b *Backend) IsIP(ip string) bool {
	host := strings.Split(b.Endpoint.Host, ":")[0]
	if host == net.ParseIP(ip).String() {
		return true
	}
	return false
}

// NewBackend generates a backend for frontend usage
func NewBackend(endpoint *url.URL, fe *Frontend) (be *Backend) {
	server := NewServer(ipfromurl(endpoint))
	be = &Backend{
		Endpoint: endpoint,
		Frontend: fe,
	}
	server.AddBackend(be)
	return
}
