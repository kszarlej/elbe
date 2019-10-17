package main

import (
	"log"
	"net"

	// "io"
	// "bufio"

	"io"
	"time"
)

const (
	LISTEN_HOST         = "localhost"
	LISTEN_PORT         = "8085"
	PROXY_PASS_HOST     = "localhost"
	PROXY_PASS_PORT     = "8082"
	CLIENT_READ_TIMEOUT = time.Microsecond * time.Duration(1000)
)

func main() {
	var config Config = readConfig()

	listener, err := net.Listen("tcp", LISTEN_HOST+":"+LISTEN_PORT)
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		backend := initConnect()

		go proxy(conn, backend, &config)
	}
}

// readMessage reads a message from socket
func readMessage(conn net.Conn, timeout time.Duration) ([]byte, error) {
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 256)

	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		num, err := conn.Read(tmp)

		if err == io.EOF || num == 0 {
			break
		}

		buf = append(buf, tmp[:num]...)
	}

	return buf, nil
}

func writeMessage(conn net.Conn, timeout time.Duration, message []byte) error {
	conn.SetWriteDeadline(time.Now().Add(timeout))
	conn.Write(message)
	return nil
}

func initConnect() net.Conn {
	conn, err := net.Dial("tcp", PROXY_PASS_HOST+":"+PROXY_PASS_PORT)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func proxy(client net.Conn, backend net.Conn, config *Config) {
	var loc *Location

	defer backend.Close()
	defer client.Close()

	request, err := readMessage(client, CLIENT_READ_TIMEOUT)

	if err != nil {
		log.Println(err)
	}

	// Get the parsed representation of the HTTP request
	httpRequestParsed := httpRequestParse(request)

	// Get the location config
	loc = locationMatcher(config.Locations, httpRequestParsed.uri)

	if httpRequestParsed.err != nil {
		client.Write(HTTP400(&httpRequestParsed))
		return
	}

	// Serialize the modified request to text format which will be sent to Client
	proxyRequest := httpMessageSerialize(httpRequestParsed)

	// Send request to upstream
	write_timeout := configGetValue(config, loc, "proxy_write_timeout").(time.Duration)
	err = writeMessage(backend, write_timeout, proxyRequest)

	// Read message from upstream
	// read_timeout := configGetValue(config, loc, "proxy_read_timeout").(time.Duration)
	response, err := readMessage(backend, CLIENT_READ_TIMEOUT)

	if isTimeout(err) {
		client.Write(HTTP504(&httpRequestParsed))
		return
	}

	// Get the parsed representation of the response
	responseParsed := httpRequestParse(response)

	// Run the proxy pipeline
	err = proxy_pipeline(&responseParsed, loc)

	if err != nil {
		log.Println(err)
	} else {
		client.Write(httpMessageSerialize(responseParsed))
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
