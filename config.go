package main

import (
	"log"

	"gopkg.in/yaml.v2"
)

const (
	CONFIG_PATH = "config.yml"
)

var data = `
locations:
  - prefix: /test
    proxy_set_header:
      - Test3 Test3Header
      - Test1 Test1Header
      - Test2 Test2Header
    proxy_hide_header:
      - Date
  - prefix: /
`

type Location struct {
	Prefix            string
	Proxy_set_header  []string
	Proxy_hide_header []string
	Proxy_set_body    string
}

type Config struct {
	Locations []Location
}

func readConfig() Config {
	Cfg := Config{}
	err := yaml.Unmarshal([]byte(data), &Cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return Cfg
}
