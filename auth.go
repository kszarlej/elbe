package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"encoding/base64"
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

	hash := strings.Split(authorizationHeader, " ")
	decoded, err := base64.StdEncoding.DecodeString(hash[1])
	if err != nil {
		fmt.Println(err)
	}
	
	creds := strings.Split(string(decoded), ":")
	username := creds[0]
	password := creds[1]

	err = bcrypt.CompareHashAndPassword([]byte(ac.BasicAuthUsers[username]), []byte(password))

	if err != nil {
		return false, &authError{err: "Bad credentials", header: false}
	}
	return true, nil
}
