package adminserver

import (
	"encoding/json"
	"fmt"
	"github.com/jeffjenkins/dispatchd/server"
	"github.com/rcrowley/go-metrics"
	"net/http"
	"os"
)

func homeJSON(w http.ResponseWriter, r *http.Request, server *server.Server) {
	var b, err = json.MarshalIndent(server, "", "    ")
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Write(b)
}

func statsJSON(w http.ResponseWriter, r *http.Request, server *server.Server) {
	// fmt.Println(metrics.DefaultRegistry)
	var b, err = json.MarshalIndent(metrics.DefaultRegistry, "", "    ")
	if err != nil {
		w.Write([]byte(err.Error()))
	}
	w.Write(b)
}

func StartAdminServer(server *server.Server, port int) {
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
	fmt.Printf("Admin server on port %d, static files from: %s\n", port, path)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
