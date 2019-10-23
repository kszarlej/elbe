package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	response = "response"
	request  = "request"
)

var (
	generalHeadersList = []string{
		"Cache-Control",
		"Connection",
		"Date",
		"Pragma",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
		"Via",
		"Warning",
	}

	requestHeadersList = []string{
		"Accept",
		"Accept-Charset",
		"Accept-Encoding",
		"Accept-Language",
		"Authorization",
		"Expect",
		"From",
		"Host",
		"If-Match",
		"If-Modified-Since",
		"If-None-Match",
		"If-Range",
		"If-Unmodified-Since",
		"Max-Forwards",
		"Proxy-Authorization",
		"Range",
		"Referer",
		"TE",
		"User-Agent",
	}

	entityHeadersList = []string{
		"Allow",
		"Content-Encoding",
		"Content-Language",
		"Content-Length",
		"Content-Location",
		"Content-MD5",
		"Content-Range",
		"Content-Type",
		"Expires",
		"Last-Modified",
	}

	allowedMethods = []string{
		"GET",
		"POST",
		"HEAD",
	}

	httpVersions = []string{
		"HTTP/1.0",
		"HTTP/1.1",
	}

	CRLF          = []byte{13, 10}
	CRLF_S        = string(CRLF)
	DOUBLE_CRLF   = []byte{13, 10, 13, 10}
	DOUBLE_CRLF_S = string(DOUBLE_CRLF)
)

// Base HTTPRequest which represents HTTP request
type HTTPMessage struct {
	direction string
	method    string
	uri       string
	version   string
	gheaders  map[string]string
	rheaders  map[string]string
	eheaders  map[string]string
	body      []byte
	err       error
	code      int
	message   string
}

type header struct {
	hname string
	hval  string
}

func (message HTTPMessage) SerializeHeaders() []byte {
	var serialized []byte

	serialize := func(headers map[string]string) {
		for headerName, headerValue := range headers {
			h := []byte(fmt.Sprintf("%s: %s%s", headerName, headerValue, CRLF_S))
			serialized = append(serialized, h[:]...)
		}
	}

	serialize(message.gheaders)
	serialize(message.rheaders)
	serialize(message.eheaders)

	return serialized
}

func (message HTTPMessage) Serialize() []byte {
	var serialized []byte

	if message.direction == request {
		serialized = []byte(fmt.Sprintf("%s %s %s%s", message.method, message.uri, message.version, CRLF_S))
	} else if message.direction == response {
		serialized = []byte(fmt.Sprintf("%s %d %s%s", message.version, message.code, message.message, CRLF_S))
	}

	serialized = append(serialized, message.SerializeHeaders()...)
	serialized = append(serialized, CRLF[:]...)
	serialized = append(serialized[:], message.body[:]...)
	return serialized
}

func allowedMethodsRegex() string {
	return fmt.Sprintf("(%s)", strings.Join(allowedMethods, "|"))
}

func httpVersionsRegex() string {
	return fmt.Sprintf("(%s)", strings.Join(httpVersions, "|"))
}

func httpReadMessage(conn net.Conn, timeout time.Duration) HTTPMessage {
	buf := make([]byte, 0, 4096)
	tmp := make([]byte, 256)
	bodytmp := make([]byte, 0, 4096)

	var headers []byte
	var httpObj HTTPMessage

	// Loop reading from the socket until DOUBLE_CRLF is found
	// DOUBLE_CRLF splits HTTP Headers from HTTP Body
	for {
		conn.SetReadDeadline(time.Now().Add(timeout))
		num, err := conn.Read(tmp)

		if err == io.EOF || num == 0 {
			break
		}

		buf = append(buf, tmp[:num]...)

		index := strings.Index(string(buf), string(DOUBLE_CRLF))

		if index > 0 {
			headers = buf[0:index]
			bodytmp = append(bodytmp, buf[index+4:len(buf)]...)
			break
		}

		httpObj.err = errors.New("Bad Request")
		return httpObj
	}

	// Parse the headers
	httpObj = httpParseHeaders(headers, httpObj)
	if httpObj.err != nil {
		return httpObj
	}

	// TODO: Learn the "comma ok" idiom
	// If Content-Length is preset then read the message body
	// Part of the body was read above when reading headers
	// so we need to calculate how much of the content is
	// still left to be read on the socket.
	if contentLength, headerPresent := httpObj.eheaders["Content-Length"]; headerPresent {
		contentLength, _ := strconv.Atoi(contentLength)
		alreadyRead := len(bodytmp)

		body := make([]byte, 0, contentLength)
		body = append(body, bodytmp...)

		if alreadyRead < contentLength {
			leftToRead := contentLength - alreadyRead

			// Keep track how many bytes are read
			readBytes := 0

			for {
				conn.SetReadDeadline(time.Now().Add(timeout))
				num, err := conn.Read(tmp)

				readBytes += num

				if err == io.EOF || num == 0 || readBytes >= leftToRead {
					break
				}

				body = append(body, tmp[:num]...)
			}
		}

		httpObj.body = body
	}

	return httpObj
}

func httpParseHeaders(headers []byte, obj HTTPMessage) HTTPMessage {
	var method, uri, version, message, direction string
	var code int

	var gheaders = make(map[string]string)
	var rheaders = make(map[string]string)
	var eheaders = make(map[string]string)

	data := strings.Split(string(headers), string(CRLF))

	// parse starting line
	starting_line := strings.Split(data[0], " ")

	var requestLine = regexp.MustCompile(fmt.Sprintf("^%s", allowedMethodsRegex()))
	var responseLine = regexp.MustCompile(fmt.Sprintf("^%s", httpVersionsRegex()))

	if requestLine.MatchString(data[0]) {
		direction = request
		method, uri, version = starting_line[0], starting_line[1], starting_line[2]
	} else if responseLine.MatchString(data[0]) {
		httpCode, _ := strconv.Atoi(starting_line[1])
		version, code, message = starting_line[0], httpCode, starting_line[2]
		direction = response
	}

	// range data[1:] to omit the request line parsed above
	for _, line := range data[1:] {
		line_split := strings.Split(line, ":")
		header_name := line_split[0]
		header_value := strings.Trim(strings.Join(line_split[1:], ":"), " ")

		if StringInSlice(header_name, generalHeadersList) {
			gheaders[header_name] = header_value
		}

		if StringInSlice(header_name, requestHeadersList) {
			rheaders[header_name] = header_value
		}

		if StringInSlice(header_name, entityHeadersList) {
			eheaders[header_name] = header_value
		}
	}

	obj.version = version

	if direction == "request" {
		if !StringInSlice(method, allowedMethods) {
			obj.err = errors.New("Unallowed method")
			return obj
		}
	}

	obj.method = method
	obj.uri = uri
	obj.gheaders = gheaders
	obj.rheaders = rheaders
	obj.eheaders = eheaders
	obj.code = code
	obj.message = message
	obj.direction = direction
	return obj
}

func HTTP400(request *HTTPMessage) []byte {
	var obj = HTTPMessage{
		version: (*request).version,
		message: "Bad Request",
		body: []byte(`<html>
<head><title>400 Bad Request</title></head>
<body>
<center><h1>400 Bad Request</h1></center>
<hr><center>elbe</center>
</body>
</html>`),
		code: 400,
		rheaders: map[string]string{
			"Connection": "close",
		},
	}

	return obj.Serialize()
}

func HTTP504(request *HTTPMessage) []byte {
	var obj = HTTPMessage{
		version: (*request).version,
		message: "Gateway Timeout",
		body: []byte(`<html>
<head><title>504 Gateway Timeout</title></head>
<body>
<center><h1>504 Gateway Timeout</h1></center>
<hr><center>elbe</center>
</body>
</html>`),
		code: 504,
		rheaders: map[string]string{
			"Connection": "close",
		},
	}

	return obj.Serialize()
}
