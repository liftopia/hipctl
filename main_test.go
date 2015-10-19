package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/soveran/redisurl"
	"github.com/stvp/tempredis"
)

func say(a ...interface{}) (n int, err error) {
	if testing.Verbose() {
		n, err = fmt.Println(a...)
	}
	return
}

func clear() {
	conn.Do("FLUSHALL")
}

func TestUsingTempRedis(t *testing.T) {
	defer clear()
	conn.Do("SET", "TESTKEY", "thething")
	value, err := redis.String(conn.Do("GET", "TESTKEY"))
	if err != nil {
		t.Error(err)
	} else {
		t.Log(value)
	}
}

func TestRedisEmpty(t *testing.T) {
	defer clear()
	_, err := redis.String(conn.Do("GET", "TESTKEY"))
	if err != nil {
		t.Log(err)
	} else {
		t.Error("redis isn't clearing between tests :(")
	}
}

func TestAddFrontend(t *testing.T) {
	defer clear()
}

func TestAddBackendByIP(t *testing.T) {
	defer clear()
	var options = []string{"testing.com", "www.testing.com"}
	fe := NewFrontend("frontend:testing.com", options, 80)
	fe.addbackend("10.10.10.10")
	if err := fe.Save(); err != nil {
		t.Error(err)
	}
}

func TestAddBadBackendByIP(t *testing.T) {
	defer clear()
	var options = []string{"testing.com", "www.testing.com"}
	fe := NewFrontend("frontend:testing.com", options, 80)
	fe.addbackend("10.10.10.1000")
	if err := fe.Save(); err != nil {
		t.Error(err)
	}
}

func TestMain(m *testing.M) {
	server, err := tempredis.Start(tempredis.Config{"databases": "8"})
	if err != nil {
		panic(err)
	}
	defer server.Term()

	server.WaitFor(tempredis.Ready)

	conn, err = redisurl.ConnectToURL(server.Config.URL().String())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	os.Exit(m.Run())
}
