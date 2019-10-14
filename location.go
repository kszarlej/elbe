package main

import (
	"errors"
	"strings"
)

func locationMatcher(locations []Location, uri string) *Location {
	var exactMatch bool = false
	var prefixes []Location
	var matched *Location = nil

	for _, location := range locations {
		// If there is an exact match then we use it
		if location.Prefix == uri {
			matched = &location
			exactMatch = true
			break
		}

		// Build a slice with all locations which are
		// prefixes to the provided URI. Later on the longest
		// prefix will be chosen.
		if strings.HasPrefix(location.Prefix, uri) {
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
func getLocationWithLongestPrefix(locations []Location) *Location {
	longest := &locations[0]
	length := len(locations[0].Prefix)

	for _, location := range locations {
		if len(location.Prefix) > length {
			longest, length = &location, len(location.Prefix)
		}
	}

	return longest
}

func getRootLocation(locations []Location) (*Location, error) {
	for _, location := range locations {
		if location.Prefix == "/" {
			return &location, nil
		}
	}

	return nil, errors.New("Root location not found")
}
