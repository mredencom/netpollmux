package mux

import (
	"log"
	"net/http"
	"testing"
)

func TestProxy(t *testing.T) {
	go func() {
		m := NewRouter()
		m.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello from 8081"))
		}).All()
		log.Fatal(m.Run(":8081"))
	}()
	m := NewRouter()
	m.HandleFunc("/proxy", func(w http.ResponseWriter, r *http.Request) {
		Proxy(w, r, "http://localhost:8081/hello")
	}).All()
	log.Fatal(m.Run(":8080"))
}
