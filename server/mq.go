package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
)

func handleConnection(server *Server, conn net.Conn) {
	server.openConnection(conn)
}

func main() {
	runtime.SetBlockProfileRate(1)
	var server = NewServer("mq.db")
	ln, err := net.Listen("tcp", ":1111")
	if err != nil {
		fmt.Printf("Error!\n")
		os.Exit(1)
	}
	fmt.Printf("Listening on port 1111\n")
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	go func() {
		startAdminServer(server)
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
