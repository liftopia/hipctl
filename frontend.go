package main

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/garyburd/redigo/redis"
)

type frontend struct {
	name       string
	id         string
	hostheader string
	port       string
	Backends   []*Backend
}

func (f *frontend) String() string {
	return fmt.Sprintf(
		"%-40s %-39s %-39s",
		f.name,
		f.id,
		f.hostheader,
	)
}

func (f *frontend) hasbackend(ip string) bool {
	for _, be := range f.Backends {
		if be.IP.String() == net.ParseIP(ip).String() {
			return true
		}
	}
	return false
}

func (f *frontend) getbackend(ip string) Backend {
	for _, be := range f.Backends {
		if be.IP.String() == net.ParseIP(ip).String() {
			return *be
		}
	}
	return Backend{}
}

func (f *frontend) addbackend(ip string) {
	fe := *f
	be := NewBackend(ip, fe.port, &fe)
	conn.Do("RPUSH", f.name, be.Endpoint)
	fe, _ = getfrontend(f.name)
	*f = fe
}

func (f *frontend) removebackend(ip string) {
	fe := *f
	be := fe.getbackend(ip)
	conn.Do("LREM", f.name, 0, be.Endpoint)
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
		Backends:   make([]*Backend, 0, len(hosts)),
	}

	for h := range hosts {
		host := hosts[h][7 : len(hosts[h])-3]
		be := NewBackend(host, fe.port, &fe)
		fe.Backends = append(fe.Backends, &be)
	}

	return fe, nil
}

func getfrontendkeys() (keys []string) {
	values, _ := redis.Values(conn.Do("KEYS", "frontend:*"))
	if len(values) == 0 {
		return nil
	}

	if err := redis.ScanSlice(values, &keys); err != nil {
		return nil
	}

	return
}

func getfrontends() (frontends map[string]frontend, err error) {
	keys := getfrontendkeys()
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
