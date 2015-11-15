package main

import (
	"encoding/json"
	"fmt"
	"github.com/rcrowley/go-metrics"
	"net/http"
	"os"
	"sort"
)

func homeJSON(w http.ResponseWriter, r *http.Request, server *Server) {
	var b, err = json.MarshalIndent(server, "", "    ")
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Write(b)
}

func statsJSON(w http.ResponseWriter, r *http.Request, server *Server) {
	// fmt.Println(metrics.DefaultRegistry)
	var b, err = json.MarshalIndent(metrics.DefaultRegistry, "", "    ")
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Write(b)
}

func home(w http.ResponseWriter, r *http.Request, server *Server) {
	// Exchange info
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<h1>Exchanges</h1>")
	fmt.Fprintf(w, "<pre>")

	var keys = make([]string, 0, len(server.exchanges))
	for name, _ := range server.exchanges {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		exchange := server.exchanges[name]
		fmt.Fprintf(
			w,
			"'%s': type: %s, bindings: %d\n",
			name,
			exchangeTypeToName(exchange.extype),
			len(exchange.bindings),
		)
	}
	fmt.Fprintf(w, "</pre>")

	// Queue info
	fmt.Fprintf(w, "<h1>Queues</h1>")
	fmt.Fprintf(w, "<pre>")
	for name, queue := range server.queues {
		fmt.Fprintf(
			w,
			"'%s': consumers: %d, queue length: %d, total received messages: %d\n",
			name,
			len(queue.consumers),
			queue.queue.Len(),
			queue.statCount,
		)
		fmt.Fprintf(w, "<h3>Consumers</h3>")
		for _, consumer := range queue.consumers {
			fmt.Fprintf(w, "'%s': %d\n", consumer.consumerTag, consumer.statCount)
		}
	}
	fmt.Fprintf(w, "</pre>")
}

func startAdminServer(server *Server) {
	// Static files
	var path = os.Getenv("STATIC_PATH")
	if len(path) == 0 {
		panic("No static file path in $STATIC_PATH!")
	}
	var fileServer = http.FileServer(http.Dir(path))
	http.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Home
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var p = path + "/admin.html"
		fmt.Println(p)
		http.ServeFile(w, r, p)
	})

	// API
	http.HandleFunc("/api/server", func(w http.ResponseWriter, r *http.Request) {
		homeJSON(w, r, server)
	})

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		statsJSON(w, r, server)
	})

	// Boot admin server
	fmt.Printf("Admin server on port 8080, static files from: %s\n", path)
	http.ListenAndServe(":8080", nil)
}
