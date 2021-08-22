package mux

import (
	"log"
	"net/http"
	"testing"
)

func TestSetHeader(t *testing.T) {
	t.Log("start...")
	m := NewRoute()
	m.Use(func(w http.ResponseWriter, r *http.Request) {
		SetHeader(w, AccessControlAllowOrigin, "*")
	})
	m.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	}).All()
	log.Fatal(http.ListenAndServe(":8080", m))
}
