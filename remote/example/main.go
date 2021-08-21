package main

import (
	"fmt"
	"log"
	"net/http"
	"netpollmux/mux"
	"netpollmux/remote"
)

func main() {
	m := mux.NewRouter()
	m.HandleFunc("/remote", func(w http.ResponseWriter, r *http.Request) {
		addr := remote.RemoteAddr(r)
		fmt.Println("addr:", addr)
		w.Write([]byte("addr:"+addr))
	}).GET()
	log.Fatal(m.Run(":8088"))
}
