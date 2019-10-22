package main

import (
	"log"
	"net"

	// "io"
	// "bufio"

	"time"
)

const (
	LISTEN_HOST         = "localhost"
	LISTEN_PORT         = "8085"
	PROXY_PASS_HOST     = "localhost"
	PROXY_PASS_PORT     = "8082"
	CLIENT_READ_TIMEOUT = time.Second * time.Duration(30)
)

func main() {
	var config Config = readConfig()

	listener, err := net.Listen("tcp", LISTEN_HOST+":"+LISTEN_PORT)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	// First call initializes the dynamic upstreams before we start the goroutine
	// that modifies it throughout the elbe lifetime. Removing this first call
	// might cause `nil` pointer dereference if immediately after program starts
	// a new connection from client will popup and first iteration of loop in goroutine
	// won't set the initial upstreams.
	SetDynamicUpstreams(&config, true)
	go SetDynamicUpstreams(&config, false)

	// Start the Accept/Proxy loop
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go proxy(conn, &config)
	}
}

// writeMessage to a socket
func writeMessage(conn net.Conn, timeout time.Duration, message []byte) error {
	conn.SetWriteDeadline(time.Now().Add(timeout))
	conn.Write(message)
	return nil
}

func initConnect(addr string) net.Conn {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func proxy(client net.Conn, config *Config) {
	var loc *Location

	defer client.Close()

	request := httpReadMessage(client, CLIENT_READ_TIMEOUT)
	if request.err != nil {
		client.Write(HTTP400(&request))
		return
	}

	// Get the location config
	loc = locationMatcher(config.Locations, request.uri)

	upstreamName := configGetValue(config, loc, "proxy_pass")
	upstreamHost := RoundRobinGetHost(upstreamName.(string))

	backend := initConnect(upstreamHost)
	defer backend.Close()

	// Serialize the modified request to text format which will be sent to Client
	serializedRequest := httpMessageSerialize(request)

	// Send request to upstream
	write_timeout := configGetValue(config, loc, "proxy_write_timeout").(time.Duration)
	writeMessage(backend, write_timeout, serializedRequest)

	// Read message from upstream
	read_timeout := configGetValue(config, loc, "proxy_read_timeout").(time.Duration)
	response := httpReadMessage(backend, read_timeout)

	// if isTimeout(err) {
	// 	client.Write(HTTP504(&httpRequestParsed))
	// 	return
	// }

	// Run the proxy pipeline
	err := proxy_pipeline(&response, loc)

	if err != nil {
		log.Println(err)
	} else {
		client.Write(httpMessageSerialize(response))
	}
}

func proxy_pipeline(message *HTTPMessage, location *Location) error {
	if len(location.Proxy_set_header) > 0 {
		proxySetHeader(message, location.Proxy_set_header)
	}

	if len(location.Proxy_hide_header) > 0 {
		proxyHideHeader(message, location.Proxy_hide_header)
	}

	return nil
}

func isTimeout(err error) bool {
	switch err := err.(type) {
	case net.Error:
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return true
		}
	}

	return false
}
