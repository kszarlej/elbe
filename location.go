package main

import (
	"errors"
	"strings"
)

type location struct {
	prefix           string
	proxy_set_header []header
}

func locationMatcher(locations []location, uri string) *location {
	var exactMatch bool = false
	var prefixes []location
	var matched *location = nil

	for _, location := range locations {
		// If there is an exact match then we use it
		if location.prefix == uri {
			matched = &location
			exactMatch = true
			break
		}

		// Build a slice with all locations which are
		// prefixes to the provided URI. Later on the longest
		// prefix will be chosen.
		if strings.HasPrefix(location.prefix, uri) {
			prefixes = append(prefixes, location)
		}
	}

	if len(prefixes) > 0 && exactMatch == false {
		matched = getLocationWithLongestPrefix(prefixes)
	}

	// assuming that first location will always be "/"
	if matched == nil {
		matched, _ = getRootLocation(locations)
	}

	return matched
}

// Returns the longest string in slice of strings provided as argument
func getLocationWithLongestPrefix(locations []location) *location {
	longest := &locations[0]
	length := len(locations[0].prefix)

	for _, location := range locations {
		if len(location.prefix) > length {
			longest, length = &location, len(location.prefix)
		}
	}

	return longest
}

func getRootLocation(locations []location) (*location, error) {
	for _, location := range locations {
		if location.prefix == "/" {
			return &location, nil
		}
	}

	return nil, errors.New("Root location not found")
}
