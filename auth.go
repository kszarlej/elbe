package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

type AuthConfig struct {
	AuthType       string `yaml:"type"`
	Passwdfile     string
	BasicAuthUsers map[string]string
}

// Authenticate receives the Authorization header content. Extracts the username/password
// and compares it to the password from AuthConfig for location.
func (ac *AuthConfig) Authenticate(authorizationHeader string) error {
	if authorizationHeader == "" {
		return errors.New("No authorization header")
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
		return errors.New("Wrong credentials")
	}
	return nil
}
