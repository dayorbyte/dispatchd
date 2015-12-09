package main

import (
	"flag"
	"fmt"
	"net"
	// _ "net/http/pprof" // uncomment for debugging
	"os"
	"path/filepath"
	"runtime"
)

func handleConnection(server *Server, conn net.Conn) {
	server.openConnection(conn)
}

func main() {
	flag.Parse()
	config := configure()
	runtime.SetBlockProfileRate(1)
	serverDbPath := filepath.Join(persistDir, "dispatchd-server.db")
	msgDbPath := filepath.Join(persistDir, "messages.db")
	var server = NewServer(serverDbPath, msgDbPath, config["users"].(map[string]interface{}))
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", amqpPort))
	if err != nil {
		fmt.Printf("Error!\n")
		os.Exit(1)
	}
	fmt.Printf("Listening on port %d\n", amqpPort)
	go func() {
		startAdminServer(server, adminPort)
	}()
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection!\n")
			os.Exit(1)
		}
		go handleConnection(server, conn)
	}
}
