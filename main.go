package main

import (
	"log"
	"net"
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
	defer client.Close()

	request := httpReadMessage(client, CLIENT_READ_TIMEOUT)
	if request.err != nil {
		client.Write(HTTP400(&request))
		return
	}

	// Get the location config
	loc := locationMatcher(config.Locations, request.uri)

	proxy_auth_pipeline(&request, loc)
	proxy_request_pipeline(&request, loc)

	if request.authenticated != true {
		client.Write(HTTP401(&request))
		return
	}

	if request.err != nil {
		client.Write(HTTP400(&request))
		return
	}

	// Get upstream name and loadbalancer which returns
	// hosts to connect.
	upstreamName := config.Get(loc, "proxy_pass")
	loadBalancer := DynamicUpstreams[upstreamName.(string)].LoadBalancer

	backend := initConnect(loadBalancer.GetHost())
	defer backend.Close()

	// Send request to upstream
	writeTimeout := config.Get(loc, "proxy_write_timeout").(time.Duration)
	writeMessage(backend, writeTimeout, request.Serialize())

	// Read message from upstream
	readTimeout := config.Get(loc, "proxy_read_timeout").(time.Duration)
	response := httpReadMessage(backend, readTimeout)

	// if isTimeout(err) {
	// 	client.Write(HTTP504(&httpRequestParsed))
	// 	return
	// }

	proxy_response_pipeline(&response, loc)

	if response.err != nil {
		client.Write(HTTP400(&response))
		return
	}

	client.Write(response.Serialize())
}

func proxy_response_pipeline(message *HTTPMessage, location *Location) {
	var err error = nil

	if len(location.Proxy_set_header) > 0 {
		err = message.SetHeaders(location.Proxy_set_header)
	}

	if len(location.Proxy_hide_header) > 0 {
		err = message.HideHeaders(location.Proxy_hide_header)
	}

	if err != nil {
		message.err = err
	}
}

func proxy_request_pipeline(message *HTTPMessage, location *Location) {
	var err error = nil

	if location.Proxy_set_body != "" {
		err = message.SetBody(location.Proxy_set_body)
	}

	if err != nil {
		message.err = err
	}
}

// proxy_auth_pipeline checks if the request should be authenticated and present the challenge if needed.
func proxy_auth_pipeline(message *HTTPMessage, loc *Location) {

	switch loc.Auth.AuthType {
	case "basic":
		err := loc.Auth.Authenticate(message.rheaders["Authorization"])

		if err != nil {
			message.authenticated = false
		}

	default:
		message.authenticated = true
	}

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
