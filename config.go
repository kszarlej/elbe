package main

import (
	"fmt"
	"log"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	CONFIG_PATH = "config.yml"
)

var data = `
proxy_read_timeout: 60
proxy_write_timeout: 60
upstreams:
    floki:
        hosts:
        - localhost:9094
        - localhost:9095
locations:
  - prefix: /test
    proxy_set_header:
      - Test3 Test3Header
      - Test1 Test1Header
      - Test2 Test2Header
    proxy_hide_header:
      - Date
    proxy_write_timeout: 5
    proxy_read_timeout: 5
    proxy_pass: floki
    proxy_set_body: test_proxy_set_body
  - prefix: /
    proxy_pass: floki
`

const (
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
}

type Config struct {
	Locations           []Location
	Upstreams           map[string]Upstream
	Proxy_read_timeout  int
	Proxy_write_timeout int
}

func readConfig() Config {
	Cfg := Config{}
	err := yaml.Unmarshal([]byte(data), &Cfg)

	configSetDefault(&Cfg.Proxy_read_timeout, PROXY_READ_TIMEOUT)
	configSetDefault(&Cfg.Proxy_write_timeout, PROXY_WRITE_TIMEOUT)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return Cfg
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

func configGetValue(config *Config, location *Location, directive string) interface{} {
	switch directive {
	case "proxy_write_timeout":
		if location.Proxy_write_timeout == 0 {
			return time.Second * time.Duration(config.Proxy_write_timeout)
		} else {
			return time.Second * time.Duration(location.Proxy_write_timeout)
		}
	case "proxy_read_timeout":
		if location.Proxy_read_timeout == 0 {
			return time.Second * time.Duration(config.Proxy_read_timeout)
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
