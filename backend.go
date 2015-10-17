package main

import (
	"fmt"
	"net"
	"net/url"
)

// Backend stores the details about the frontend's connections
type Backend struct {
	IP       net.IP
	Endpoint *url.URL
	Frontend *frontend
}

func (b *Backend) String() string {
	return fmt.Sprintf(
		"%-40s %s",
		b.IP,
		b.Endpoint,
	)
}

// NewBackend generates a backend for frontend usage
func NewBackend(ip string, port string, fe *frontend) Backend {
	return Backend{
		Endpoint: &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%s", ip, port),
		},
		Frontend: fe,
		IP:       net.ParseIP(ip),
	}
}
