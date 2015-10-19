package main

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

func urlfromipandport(ip net.IP, port int) *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", ip, port),
	}
}

func hostfromurl(url *url.URL) string {
	return strings.Split(url.Host, ":")[0]
}

func ipfromurl(url *url.URL) net.IP {
	return net.ParseIP(hostfromurl(url))
}
