package main

import (
	"bytes"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	name     string
	password []byte
}

func (s *Server) addUsers(userJson map[string]interface{}) {
	if len(userJson) == 0 {
		decoded, err := base64.StdEncoding.DecodeString(defaultUserPasswordBase64)
		if err != nil {
			panic("System integrity error: Could not base64 decode password for built in default user!")
		}
		s.users[defaultUserName] = User{name: defaultUserName, password: decoded}
	}
	for name, user := range userJson {
		var encoded = user.(map[string]interface{})["password_bcrypt_base64"].(string)
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			panic("Could not base64 decode password for default config file user: " + name)
		}
		s.users[name] = User{name: name, password: decoded}
	}
}

// guest/guest
var defaultUserName = "guest"
var defaultUserPasswordBase64 = "JDJhJDExJENobGk4dG5rY0RGemJhTjhsV21xR3VNNnFZZ1ZqTzUzQWxtbGtyMHRYN3RkUHMuYjF5SUt5"

func (s *Server) authenticate(mechanism string, blob []byte) bool {
	// Split. SASL PLAIN has three parts
	parts := bytes.Split(blob, []byte{0})
	if len(parts) != 3 {
		return false
	}

	for name, user := range s.users {
		if string(parts[1]) != name {
			continue
		}
		err := bcrypt.CompareHashAndPassword(user.password, parts[2])
		if err == nil {
			return true
		}
	}
	return false
}
