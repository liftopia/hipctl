package main

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/garyburd/redigo/redis"
)

// Frontend stores the details about the frontends
type Frontend struct {
	name     string
	key      string
	options  []string
	port     int
	Backends map[url.URL]*Backend
}

var frontends map[string]Frontend

func init() {
	frontends = make(map[string]Frontend)
}

func (f *Frontend) String() string {
	return fmt.Sprintf(
		"%-40s %-6d %+v",
		f.key,
		f.port,
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

// ShowFrontend gives detailed information about a frontend
func ShowFrontend(name string) {
	for _, f := range frontends {
		if f.name == name {
			fmt.Printf("%20s: %s\n", "Name", f.name)
			fmt.Printf("%20s: %s\n", "Key", f.key)
			fmt.Printf("%20s: %s\n", "Options", f.options)
			fmt.Printf("%20s: %d\n", "Port", f.port)
			fmt.Printf("%20s: %p\n", "*Address", &f)

			for endpoint, backend := range f.Backends {
				fmt.Println()
				fmt.Printf("  %p %20s: %+v\n", &endpoint, "[Endpoint]", endpoint)
				backend.Show()
			}
			return
		}
	}
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
	f.Backends[*b.Endpoint] = b
}

func (f *Frontend) addbackend(arg string) (err error) {
	ip := net.ParseIP(arg)
	be := NewBackend(urlfromipandport(ip, f.port), f)
	be.Server = NewServer(ip)
	log.Debug("Adding new backend %v to %s", be.Endpoint, f.key)
	f.appendBackend(be)
	log.Debug("%s backends: %+v", f.key, f.Backends)
	err = f.Save()

	return
}

func (f *Frontend) removebackend(be *Backend) error {
	delete(f.Backends, *be.Endpoint)
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

	log.Debug("Saving %s", f.key)
	log.Debug("%s %s", "RPUSH", create)
	log.Debug("%s %s", "RENAME", replace)

	conn.Send("MULTI")
	conn.Send("RPUSH", create...)
	conn.Send("RENAME", replace...)
	r, err := conn.Do("EXEC")
	log.Debug("Save response: %+v (%+v)", r, err)

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
		Backends: make(map[url.URL]*Backend),
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
		log.Debug("Port error for %s: %v", key, err)
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
		fe.Backends[*endpoint] = &be
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
		log.Error("Redis had an error :( %+v", err)
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
			log.Debug("Error loading frontend: %s (%s)", key, err)
			continue
		}
		frontends[key] = fe
	}

	return nil
}
