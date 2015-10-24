package main

import (
	"fmt"
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

var backends backendpool

// Port grabs the port from the Endpoint's Host key
func (b *Backend) Port() (port int) {
	port = 80
	if parts := strings.Split(b.Endpoint.Host, ":"); len(parts) > 1 {
		port, _ = strconv.Atoi(parts[1])
	}
	return
}

// Host grabs the host from the Endpoint's Host key
func (b *Backend) Host() (host string) { return strings.Split(b.Endpoint.Host, ":")[0] }

func (b *Backend) String() string {
	return fmt.Sprintf("%s, F: %s, S: %s", b.Endpoint, b.Frontend, b.Server)
}

// Show the backend's detailed information
func (b *Backend) Show() {
	fmt.Printf(showformat, &b.Endpoint, "Endpoint", b.Endpoint)
	fmt.Printf(showformat, &b.Frontend, "Frontend", b.Frontend)
	fmt.Printf(showformat, &b.Server, "Server", b.Server)
	fmt.Printf(showformat, &b, "<self>", b)
}

// ListBackends grabs the *entire* list of backends and prints 'em out
func ListBackends() {
	for _, f := range frontends {
		for _, b := range f.Backends {
			backends.append(b)
		}
	}
	fmt.Println(backends.list())
	fmt.Printf("%d backends\n", len(backends))
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
	if b.Host() == net.ParseIP(ip).String() {
		return true
	}
	return false
}

// NewBackend generates a backend for frontend usage
func NewBackend(endpoint *url.URL, fe *Frontend) (be *Backend) {
	be = &Backend{
		Endpoint: endpoint,
		Frontend: fe,
		Server:   NewServer(ipfromurl(endpoint)),
	}
	be.Server.AddBackend(be)
	return
}
