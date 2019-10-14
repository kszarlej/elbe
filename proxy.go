package main

import (
	"strings"
)

// TODO: Optimalize it (remove unneeded Split/Join)
func proxySetHeader(httpObject *HTTPMessage, headers []string) {
	for _, header := range headers {
		header := strings.Split(header, " ")
		httpObject.eheaders[header[0]] = strings.Join(header[1:], " ")
	}
}

func proxyHideHeader(httpObject *HTTPMessage, headers []string) {
	for _, header := range headers {
		delete(httpObject.eheaders, header)
	}
}
