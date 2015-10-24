package main

import (
	"fmt"
	"net/url"
)

type backendpool []*Backend

func (bp *backendpool) String() string { return fmt.Sprintf("%d backends", len(*bp)) }

func (bp *backendpool) newBackend(endpoint *url.URL, fe *Frontend) (be *Backend) {
	be = &Backend{
		Endpoint: endpoint,
		Frontend: fe,
		Server:   NewServer(ipfromurl(endpoint)),
	}
	be.Server.AddBackend(be)
	return
}

func (bp *backendpool) list() (r string) {
	for _, b := range *bp {
		r += fmt.Sprintf("%30s\t%40s\t%s\n", b.Endpoint, b.Frontend.name, b.Server.IP)
	}
	return
}

func (bp *backendpool) append(b *Backend) {
	*bp = append(*bp, b)
}

func (bp *backendpool) addserver(s *Server) {
	for _, b := range *bp {
		b.Server = s
		s.AddBackend(b)
	}
}
