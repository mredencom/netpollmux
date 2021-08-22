package main

import (
	"fmt"
	"github.com/php2go/netpollmux/mux"
	"log"
	"net/http"
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
