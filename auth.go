package main

import (
	_ "fmt"
)

type authError struct {
	err string
	header bool
}

func (e *authError) Error() string {
	return e.err
}

func (e *authError) HeaderPresent() bool {
	return e.header
}

type AuthConfig struct {
	AuthType   	   string `yaml:"type"`
	Passwdfile 	   string
	BasicAuthUsers map[string]string
}

func (ac AuthConfig) Authenticate(authorizationHeader string) (bool, error) {
	if authorizationHeader == "" {
		return false, &authError{err: "Authorization required", header: false}
	}

	return true, nil
}
