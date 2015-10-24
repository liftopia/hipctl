package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
)

// Frontend stores the details about the frontends
type Frontend struct {
	name     string
	key      string
	options  []string
	port     int
	Backends map[*url.URL]*Backend
}

var frontends []*Frontend

func (f *Frontend) String() string {
	return fmt.Sprintf(
		"%-40s %-6d %d backends %+v",
		f.name,
		f.port,
		len(f.Backends),
		f.options,
	)
}

// ListFrontendsComplete prints the server list for shell completions
func ListFrontendsComplete(c *cli.Context) {
	if len(c.Args()) > 0 {
		return
	}
	for _, frontend := range frontends {
		fmt.Println(frontend.name)
	}
}

// ListFrontends outputs a concise list for viewing pleasure
func ListFrontends() {
	for _, fe := range frontends {
		fmt.Printf("%v\n", fe)
	}
	fmt.Printf("%d frontends\n", len(frontends))
}

// ShowFrontend gives detailed information about a frontend
func ShowFrontend(name string) {
	for _, f := range frontends {
		if f.name == name {
			fmt.Printf(showformat, &f.name, "Name", f.name)
			fmt.Printf(showformat, &f.key, "Key", f.key)
			fmt.Printf(showformat, &f.options, "Options", f.options)
			fmt.Printf(showformat, &f.port, "Port", f.port)
			fmt.Printf(showformat, &f, "<self>", &f)
			fmt.Printf(showformat, &f.Backends, "Backends", nil)

			var backend *Backend
			for endpoint := range f.Backends {
				backend = f.Backends[endpoint]
				fmt.Println()
				fmt.Printf(showformat, &endpoint, "[Endpoint]", endpoint)
				fmt.Printf(showformat, &backend, "[Backend]", backend)
				backend.Show()
			}
			return
		}
	}

	log.Error("Can't find that frontend!")
}

func (f *Frontend) hasbackend(ip string) (hasit bool) {
	be := f.getbackend(ip)
	if be != nil {
		hasit = !be.Empty()
	}

	return
}

func (f *Frontend) getbackend(ip string) (be *Backend) {
	for k, be := range f.Backends {
		if be.IsIP(ip) {
			return f.Backends[k]
		}
	}

	return
}

func (f *Frontend) appendBackend(b *Backend) {
	f.Backends[b.Endpoint] = b
}

func (f *Frontend) addbackend(arg string) error {
	ip := net.ParseIP(arg)
	be := NewBackend(urlfromipandport(ip, f.port), f)
	log.Debugf("Adding new backend %v to %s", be.Endpoint, f.key)
	f.appendBackend(be)
	log.Debugf("%s backends: %+v", f.key, f.Backends)

	return f.Save()
}

func (f *Frontend) removebackend(be *Backend) error {
	delete(f.Backends, be.Endpoint)
	return f.Save()
}

// Save writes the current config from memory to hipache
func (f *Frontend) Save() (err error) {
	var create []interface{}

	create = append(create, fmt.Sprintf("PREP:%s", f.key))
	create = append(create, f.printoptions())
	for _, be := range f.Backends {
		create = append(create, be.Endpoint.String())
	}

	var replace []interface{}

	replace = append(replace, fmt.Sprintf("PREP:%s", f.key))
	replace = append(replace, f.key)

	log.Debugf("Saving %s", f.key)
	log.Debugf("%s %s", "RPUSH", create)
	log.Debugf("%s %s", "RENAME", replace)

	conn.Send("MULTI")
	conn.Send("RPUSH", create...)
	conn.Send("RENAME", replace...)
	r, err := conn.Do("EXEC")
	log.Debugf("Save response: %+v (%+v)", r, err)

	return
}

func getport(endpoint string) (port int, err error) {
	var uri *url.URL
	if uri, err = url.Parse(endpoint); err != nil {
		return -1, err
	}

	parts := strings.Split(uri.Host, ":")
	if len(parts) == 1 {
		// this is gonna have to be replaced by a smarter backend handling of ports
		// containerization is going to destroy this 80 vs 81 business
		return 80, errors.New("frontend is probably using external backend")
	}

	if port, err = strconv.Atoi(parts[1]); err != nil {
		return
	}

	if port != 80 && port != 81 {
		return port, errors.New("invalid port config")
	}

	return
}

func parseinfo(str string) []string {
	return strings.Split(str, "|")
}

func (f *Frontend) printoptions() string {
	return strings.Join(f.options, "|")
}

// NewFrontend creates a new frontend and adds it to hipache's configuration
// this starts with an empty backend list
func NewFrontend(key string, options []string, port int) Frontend {
	return Frontend{
		name:     strings.Split(key, ":")[1],
		key:      key,
		options:  options,
		port:     port,
		Backends: make(map[*url.URL]*Backend),
	}
}

func getfrontend(key string) (fe Frontend, err error) {
	values, err := redis.Values(conn.Do("LRANGE", key, 0, -1))
	if err != nil {
		return
	}
	if len(values) <= 1 {
		// empty: no config at all, one: info with no hosts
		return fe, fmt.Errorf("No frontend config for %s", key)
	}

	var config []string
	if err = redis.ScanSlice(values, &config); err != nil {
		return
	}
	info, hosts := config[0], config[1:]
	options := parseinfo(info)

	var port int
	if port, err = getport(hosts[0]); err != nil {
		log.Debugf("Port error for %s: %v", key, err)
		return
	}

	fe = NewFrontend(key, options, port)

	for h := range hosts {
		host := hosts[h][7 : len(hosts[h])-3]
		ip := net.ParseIP(host)
		var be Backend
		endpoint := urlfromipandport(ip, fe.port)
		be = *NewBackend(endpoint, &fe)
		if err != nil {
			return
		}
		fe.Backends[endpoint] = &be
	}

	return
}

func getfrontendkeys() (keys []string, err error) {
	values, err := redis.Values(conn.Do("KEYS", "frontend:*"))
	if len(values) == 0 {
		log.Error("Did not find any frontend keys (frontend:*)")
		return
	}
	if err != nil {
		log.Errorf("Redis had an error :( %+v", err)
		return
	}

	redis.ScanSlice(values, &keys)
	return
}

func updatefrontends() (err error) {
	var keys []string
	if keys, err = getfrontendkeys(); err != nil {
		return
	}
	for _, key := range keys {
		if strings.Contains(key, "fr-ca") || strings.Contains(key, "blog") {
			continue
		}

		fe, err := getfrontend(key)
		if err != nil {
			log.WithFields(logrus.Fields{
				"key": key,
				"err": err,
			}).Debugf("Error loading frontend: %s (%s)", key, err)
			continue
		}
		frontends = append(frontends, &fe)
	}

	return nil
}
