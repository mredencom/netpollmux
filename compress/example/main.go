package main

import (
	"log"
	"net/http"
	"netpollmux/mux"
	"netpollmux/render"
)

func main() {
	r := render.NewRender()
	r.GzipAll().DeflateAll().Charset("utf-8")
	m := mux.NewRouter()
	m.HandleFunc("/compress", func(w http.ResponseWriter, req *http.Request) {
		r.JSON(w, req, []string{"compress"}, http.StatusOK)
	}).GET().POST().HEAD()
	log.Fatal(http.ListenAndServe(":8080", m))
}
