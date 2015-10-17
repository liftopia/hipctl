package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"

	"github.com/garyburd/redigo/redis"
)

type hostmap map[string]bool

type frontend struct {
	name       string
	id         string
	hostheader string
	port       string
	hosts      hostmap
	Backends   []*backend
}

type backend struct {
	IP        net.IP
	Endpoint  *url.URL
	Frontends hostmap
}

func (a hostmap) keys() []string {
	keys := make([]string, 0, len(a))

	for k := range a {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func (f *frontend) String() string {
	return fmt.Sprintf(
		"%-40s %-39s %-39s",
		f.name,
		f.id,
		f.hostheader,
	)
}

func (b *backend) String() string {
	return fmt.Sprintf(
		"%-40s %s",
		b.IP,
		b.Endpoint,
	)
}

func (f *frontend) hasbackend(ip string) bool {
	return f.hosts[ip]
}

func (f *frontend) addbackend(ip string) {
	fe := *f
	url := fmt.Sprintf("http://%s:%s", ip, f.port)
	conn.Do("RPUSH", f.name, url)
	fe, _ = getfrontend(f.name)
	*f = fe
}

func (f *frontend) removebackend(ip string) {
	fe := *f
	url := fmt.Sprintf("http://%s:%s", ip, f.port)
	conn.Do("LREM", f.name, 0, url)
	fe, _ = getfrontend(f.name)
	*f = fe
}

func getfrontend(key string) (fe frontend, err error) {
	values, err := redis.Values(conn.Do("LRANGE", key, 0, -1))
	if err != nil || len(values) <= 1 {
		// empty: no config at all, one: info with no hosts
		return frontend{}, err
	}

	var hosts []string
	if err := redis.ScanSlice(values, &hosts); err != nil {
		return frontend{}, err
	}

	info := strings.Split(hosts[0], "|")
	id := info[0]
	var hostheader string
	if len(info) > 1 {
		hostheader = info[1]
	}
	hosts = hosts[1:]

	port := hosts[0][len(hosts[0])-2:]
	if port != "80" && port != "81" {
		return frontend{}, errors.New("invalid port config")
	}

	fe = frontend{
		name:       key,
		id:         id,
		hostheader: hostheader,
		port:       port,
		hosts:      make(map[string]bool, len(hosts)),
		Backends:   make([]*backend, 0, len(hosts)),
	}

	for h, uri := range hosts {
		endpoint, err := url.Parse(uri)
		if err != nil {
			fmt.Printf("URI wouldn't parse: uri: %s", uri)
			fmt.Println(err)
			continue
		}
		host := hosts[h][7 : len(hosts[h])-3]
		be := backend{
			Endpoint:  endpoint,
			IP:        net.ParseIP(host),
			Frontends: make(map[string]bool),
		}
		fe.hosts[host] = true
		fe.Backends = append(fe.Backends, &be)
	}

	for _, be := range fe.Backends {
		be.Frontends[fe.name] = true
	}

	return fe, nil
}

func getfrontends() (frontends map[string]frontend, err error) {
	values, err := redis.Values(conn.Do("KEYS", "frontend:*"))
	if err != nil || len(values) == 0 {
		return nil, err
	}

	var keys []string
	if err := redis.ScanSlice(values, &keys); err != nil {
		return nil, err
	}

	frontends = make(map[string]frontend)
	for _, key := range keys {
		if strings.Contains(key, "fr-ca") || strings.Contains(key, "blog") {
			continue
		}

		fe, err := getfrontend(key)
		if err != nil {
			continue
		}
		frontends[key] = fe
	}

	return frontends, nil
}
