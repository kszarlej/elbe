package main

type Authenticator interface {
	authenticate(username, hash, password string) bool
}

type BasicAuthUser struct {
	username string
	hash     string
}

func (User BasicAuthUser) bcryptAuthenticate(password string) {

}
