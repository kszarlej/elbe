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
	LISTEN_HOST          = "localhost"
	LISTEN_PORT          = "8085"
	PROXY_PASS_HOST      = "localhost"
	PROXY_PASS_PORT      = "8082"
	CLIENT_READ_TIMEOUT  = time.Microsecond * time.Duration(1000)
	BACKEND_READ_TIMEOUT = time.Second * time.Duration(10)
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

func readMessage(conn net.Conn, timeout time.Duration) ([]byte, error) {
	buf := make([]byte, 0, 4096) // TODO LEARN WHAT IS 0
	tmp := make([]byte, 256)

	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		num, err := conn.Read(tmp)
		if err != nil {
			if err != io.EOF {
				log.Println("Error when reading from connection: ", err)
			}
		}

		if err == io.EOF || num == 0 {
			break
		}

		buf = append(buf, tmp[:num]...)
	}

	return buf, nil
}

func initConnect() net.Conn {
	conn, err := net.Dial("tcp", PROXY_PASS_HOST+":"+PROXY_PASS_PORT)
	if err != nil {
		log.Fatal(err)
	}

	return conn
}

func proxy(client net.Conn, backend net.Conn, config *Config) {
	defer backend.Close()
	defer client.Close()

	request, err := readMessage(client, CLIENT_READ_TIMEOUT)
	if err != nil {
		log.Println(err)
	}

	// Get the parsed representation of the HTTP request
	httpRequestParsed := httpRequestParse(request)

	if httpRequestParsed.err != nil {
		client.Write(HTTP400(&httpRequestParsed))
	} else {
		// Get the location config
		var loc *Location = locationMatcher(config.Locations, httpRequestParsed.uri)

		// Serialize the modified request to text format.
		proxyRequest := httpMessageSerialize(httpRequestParsed)

		// Send request to upstream
		backend.Write(proxyRequest)

		// Read message fro, upstream
		response, err := readMessage(backend, BACKEND_READ_TIMEOUT)

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
