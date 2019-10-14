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

var (
	LOCATIONS = []location{
		location{prefix: "/"},
		location{
			prefix:           "/test",
			proxy_set_header: []header{{"Test", "test header"}, {"Test2", "test header2"}},
		},
		location{prefix: "/test/test1/test2"},
	}
)

func main() {
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

		go proxy(conn, backend)
	}
}

func readRequest(conn net.Conn, timeout time.Duration) ([]byte, error) {
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

func proxy(client net.Conn, backend net.Conn) {
	defer backend.Close()
	defer client.Close()

	request, err := readRequest(client, CLIENT_READ_TIMEOUT)
	if err != nil {
		log.Println(err)
	}

	httpData := httpRequestParse(request)

	if httpData.err != nil {
		client.Write(HTTP400(&httpData))
	} else {
		var loc *location = locationMatcher(LOCATIONS, httpData.uri)

		proxyRequest := httpMessageSerialize(httpData)

		backend.Write(proxyRequest)

		response, err := readRequest(backend, BACKEND_READ_TIMEOUT)

		responseParsed := httpRequestParse(response)

		proxySetHeader(&responseParsed, loc.proxy_set_header)

		if err != nil {
			log.Println(err)
		}

		client.Write(httpMessageSerialize(responseParsed))
	}
}
