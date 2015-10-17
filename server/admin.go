package main

import (
	"fmt"
	"net/http"
	"sort"
)

func startAdminServer(server *Server) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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
				queue.consumers.Len(),
				queue.queue.Len(),
				queue.statCount,
			)
			fmt.Fprintf(w, "<h3>Consumers</h3>")
			for e := queue.consumers.Front(); e != nil; e = e.Next() {
				var consumer = e.Value.(*Consumer)
				fmt.Fprintf(w, "'%s': %d\n", consumer.consumerTag, consumer.statCount)
			}
		}
		fmt.Fprintf(w, "</pre>")
	})
	fmt.Println("Admin server on port 8080")
	http.ListenAndServe(":8080", nil)
}
