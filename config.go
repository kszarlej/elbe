package main

import (
	"fmt"
	"log"
	"time"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	CONFIG_PATH         = "config.yml"
	PROXY_READ_TIMEOUT  = 60
	PROXY_WRITE_TIMEOUT = 60
	PROXY_SET_BODY      = ""
)

type Upstream struct {
	Hosts        []string
	LoadBalancer *roundrobin
}

type Location struct {
	Prefix              string
	Proxy_set_header    []string
	Proxy_hide_header   []string
	Proxy_set_body      string
	Proxy_read_timeout  int
	Proxy_write_timeout int
	Proxy_pass          string
	Auth                AuthConfig
}

type Config struct {
	Locations           []Location
	Upstreams           map[string]Upstream
	Proxy_read_timeout  int
	Proxy_write_timeout int
}

// Iterates locations and loads loadPasswdFile on each AuthType 
func (c *Config) loadAuth() {
	for idx, l := range c.Locations {
		// Fix check l.Auth.Passwdfile check
		if l.Auth.AuthType == "basic" && l.Auth.Passwdfile != "" {
			var bausers = make(map[string]string)

			contents, _ := ioutil.ReadFile(l.Auth.Passwdfile)
			entries := strings.Split(string(contents), "\n")

			for _, e := range entries {
				user := strings.Split(e, ":")
				bausers[user[0]] = user[1]
			}

			// TODO: why commented one doesn't work
			//l.Auth.BasicAuthUsers = bausers
			c.Locations[idx].Auth.BasicAuthUsers = bausers
		}
	}
}

func readConfig() Config {
	configFile, err := ioutil.ReadFile("config.yml")
	if err != nil {
		panic("Error reading config file")
	}

	cfg := Config{}
	err = yaml.Unmarshal(configFile, &cfg)

	configSetDefault(&cfg.Proxy_read_timeout, PROXY_READ_TIMEOUT)
	configSetDefault(&cfg.Proxy_write_timeout, PROXY_WRITE_TIMEOUT)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	cfg.loadAuth()

	return cfg
}

// Sets default values for config directives if they are not specified
// by user in config file.
func configSetDefault(directive interface{}, value interface{}) {
	switch v := directive.(type) {
	case (*int):
		asserted := directive.(*int)
		if (*asserted) == 0 {
			*asserted = value.(int)
		}
	case *string:
		asserted := directive.(*string)
		if (*asserted) == "" {
			*asserted = value.(string)
		}
	default:
		fmt.Printf("I don't know about type %T!\n", v)
	}
}

// TODO: Switch to method on Config object
func (c *Config) Get(location *Location, directive string) interface{} {
	switch directive {

	case "proxy_write_timeout":
		if location.Proxy_write_timeout == 0 {
			return time.Second * time.Duration(c.Proxy_write_timeout)
		} else {
			return time.Second * time.Duration(location.Proxy_write_timeout)
		}
	
	case "proxy_read_timeout":
		if location.Proxy_read_timeout == 0 {
			return time.Second * time.Duration(c.Proxy_read_timeout)
		} else {
			return time.Second * time.Duration(location.Proxy_read_timeout)
		}
	
	case "proxy_set_body":
		return location.Proxy_set_body
	
	case "proxy_pass":
		return location.Proxy_pass

	default:
		fmt.Printf("Hello :)")
	}

	return nil
}