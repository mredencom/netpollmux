package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"netpollmux"
	"netpollmux/mux"
	"netpollmux/response"
	"sync"
)
// https://github.com/hslam?tab=repositories
func main() {
	m := mux.NewRoute()
	//m := gmux.NewRouter()
	m.HandleFunc("/hello/:id", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`Hello World`))
		return
		//fmt.Fprintf(w, "Category: %v\n", params)

	})
	if err := ListenAndServe(":8080", m); err != nil {
		log.Fatal("启动失败")
	}
}

type Context struct {
	reader  *bufio.Reader
	rw      *bufio.ReadWriter
	conn    net.Conn
	serving sync.Mutex
}

func ListenAndServe(addr string, handler http.Handler) error {
	var h = netpollmux.NewConHandler()

	h.SetUpgrade(func(conn net.Conn) (netpollmux.Context, error) {
		reader := bufio.NewReader(conn)
		rw := bufio.NewReadWriter(reader, bufio.NewWriter(conn))
		return &Context{reader: reader, conn: conn, rw: rw}, nil
	})
	h.SetServe(func(context netpollmux.Context) error {
		ctx := context.(*Context)
		ctx.serving.Lock()
		req, err := http.ReadRequest(ctx.reader)
		if err != nil {
			ctx.serving.Unlock()
			return err
		}
		res := response.NewResponse(req, ctx.conn, ctx.rw)
		handler.ServeHTTP(res, req)
		res.FinishRequest()
		ctx.serving.Unlock()
		response.FreeResponse(res)
		return nil
	})
	return netpollmux.ListenAndServe("tcp", addr, h)
}
