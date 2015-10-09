package main

import (
	"fmt"
	"net"
	"os"
)

func handleConnection(server *Server, conn net.Conn) {
	server.openConnection(conn)
}

func main() {
	fmt.Printf("Listening on port 1111\n")
	var server = NewServer()
	ln, err := net.Listen("tcp", ":1111")
	if err != nil {
		fmt.Printf("Error!\n")
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
