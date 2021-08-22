package main

import (
	"github.com/php2go/netpollmux/mux"
	"log"
	"net/http"
)

func main() {
	r := mux.NewRender()
	r.GzipAll().DeflateAll().Charset("utf-8")
	m := mux.NewRouter()
	m.HandleFunc("/compress", func(w http.ResponseWriter, req *http.Request) {
		r.JSON(w, req, []string{"compress"}, http.StatusOK)
	}).GET().POST().HEAD()
	log.Fatal(http.ListenAndServe(":8080", m))
}
