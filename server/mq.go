package main

import (
	"fmt"
	"net"
	"os"
)

func handleConnection(server *Server, conn net.Conn) {
	c := NewAMQPConnection(server, conn)
	c.openConnection()
}

func main() {
	fmt.Printf("Listening on port 1111\n")
	var server = NewServer()
	ln, err := net.Listen("tcp", ":1111")
	if err != nil {
		fmt.Printf("Error!")
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection!\n")
			os.Exit(1)
		}
		go handleConnection(server, conn)
	}
}
