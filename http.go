package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

func allowedMethodsRegex() string {
	return fmt.Sprintf("(%s)", strings.Join(allowedMethods, "|"))
}

func httpVersionsRegex() string {
	return fmt.Sprintf("(%s)", strings.Join(httpVersions, "|"))
}

func httpRequestParse(httpData []byte) HTTPMessage {
	var heading []byte
	var body []byte
	var obj HTTPMessage

	// Search for first double CRLF occurance.
	// First occurance separates the HTTP message
	// headers(+request line) and body.
	index := strings.Index(string(httpData), string(DOUBLE_CRLF))

	if index == -1 {
		// HTTP request without body
		heading = httpData[0:len(httpData)]
		body = DOUBLE_CRLF
	} else {
		// HTTP request with body
		heading = httpData[0:index]
		body = httpData[index+4 : len(httpData)]
	}

	obj = httpParseHeaders(heading, obj)
	obj.body = body

	return obj
}

func httpSerializeHeaders(headers map[string]string) []byte {
	var serialized []byte
	for header_name, header_value := range headers {
		header := []byte(fmt.Sprintf("%s: %s%s", header_name, header_value, CRLF_S))
		serialized = append(serialized, header[:]...)
	}
	return serialized
}

func httpMessageSerialize(message HTTPMessage) []byte {
	var serialized []byte

	if message.direction == request {
		serialized = []byte(fmt.Sprintf("%s %s %s%s", message.method, message.uri, message.version, CRLF_S))
	} else if message.direction == response {
		serialized = []byte(fmt.Sprintf("%s %d %s%s", message.version, message.code, message.message, CRLF_S))
	}

	serialized = append(serialized, httpSerializeHeaders(message.gheaders)[:]...)
	serialized = append(serialized, httpSerializeHeaders(message.rheaders)[:]...)
	serialized = append(serialized, httpSerializeHeaders(message.eheaders)[:]...)
	serialized = append(serialized, CRLF[:]...)
	serialized = append(serialized[:], message.body[:]...)
	return serialized
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

	return httpMessageSerialize(obj)
}
