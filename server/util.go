package main

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func randomId() string {
	var size = 32
	var numChars = len(chars)
	id := make([]rune, size)
	for i := range id {
		id[i] = chars[rand.Intn(numChars)]
	}
	return string(id)
}
