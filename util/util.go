package util

import (
	"math/rand"
	"sync/atomic"
	"time"
)

var counter int64

func init() {
	rand.Seed(time.Now().UnixNano())
	// System Integrity Note:
	//
	// Start the counter at the time of server boot. This is so that we have
	// message IDs which are consistent within a single server even if it is
	// restarted. One implication of using the time is that if the server boot
	// time is earlier than the previous boot time durable messages may be
	// loaded out of order
	counter = time.Now().UnixNano()
}

var chars = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func RandomId() string {
	var size = 32
	var numChars = len(chars)
	id := make([]rune, size)
	for i := range id {
		id[i] = chars[rand.Intn(numChars)]
	}
	return string(id)
}

func NextId() int64 {
	return atomic.AddInt64(&counter, 1)
}
