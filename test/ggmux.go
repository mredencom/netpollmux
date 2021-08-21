package main

import (
	"log"
	"net/http"
	"netpollmux/gmux"
	"time"
)

func main() {
	r := gmux.NewRouter()

	// This will serve files under http://localhost:8000/static/<filename>
	r.HandleFunc("/hello/{key}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`Hello World`))
		return
	})
	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
