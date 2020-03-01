package main

import (
	"strconv"
	"strings"
)

// TODO: Optimize it (remove unneeded Split/Join)
func proxySetHeader(httpObject *HTTPMessage, headers []string) {
	for _, header := range headers {
		header := strings.Split(header, " ")
		httpObject.eheaders[header[0]] = strings.Join(header[1:], " ")
	}
}

func proxyHideHeader(httpObject *HTTPMessage, headers []string) {
	for _, header := range headers {
		delete(httpObject.eheaders, header)
		delete(httpObject.rheaders, header)
		delete(httpObject.gheaders, header)
	}
}

func proxySetBody(httpObject *HTTPMessage, body string) {
	httpObject.body = []byte(body)
	httpObject.eheaders["Content-Length"] = strconv.Itoa(len(body))
}
