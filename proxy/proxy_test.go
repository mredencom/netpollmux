package proxy

import (
	"log"
	"net/http"
	"netpollmux/mux"
	"testing"
)

func TestProxy(t *testing.T) {
	go func() {
		m := mux.NewRouter()
		m.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello from 8081"))
		}).All()
		log.Fatal(m.Run(":8081"))
	}()
	m := mux.NewRouter()
	m.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		Proxy(w, r, "http://localhost:8081/hello")
	}).All()
	log.Fatal(m.Run(":8080"))
}
