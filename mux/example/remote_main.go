package main

import (
	"fmt"
	"log"
	"net/http"
	"netpollmux/mux"
)

func main() {
	m := mux.NewRouter()
	m.HandleFunc("/remote", func(w http.ResponseWriter, r *http.Request) {
		addr := mux.RemoteAddr(r)
		fmt.Println("addr:", addr)
		w.Write([]byte("addr:" + addr))
	}).GET()
	log.Fatal(m.Run(":8088"))
}
