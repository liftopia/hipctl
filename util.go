package main

import (
	"fmt"
	"net/url"
)

func urlfromipandport(ip string, port int) url.URL {
	return url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", ip, port),
	}
}
