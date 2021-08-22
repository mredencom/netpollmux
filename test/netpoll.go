package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/php2go/netpollmux/mux"
	"github.com/php2go/netpollmux/netpoll"
)

// https://github.com/hslam?tab=repositories
func main() {
	m := mux.NewRouter()
	r := mux.NewRender()
	r.GzipAll().DeflateAll().Charset("utf-8")
	//m := gmux.NewRouter()
	m.HandleFunc("/hello/:id", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(`Hello World`))
		//r.JSON(w, req, []string{"compress"}, http.StatusOK)
	})
	log.Fatal(m.Run(":8080"))
	//if err := ListenAndServe(":8080", m); err != nil {
	//	log.Fatal("启动失败")
	//}
}

type Context struct {
	reader  *bufio.Reader
	rw      *bufio.ReadWriter
	conn    net.Conn
	serving sync.Mutex
}

func ListenAndServe(addr string, handler http.Handler) error {
	var h = netpoll.NewConHandler()

	h.SetUpgrade(func(conn net.Conn) (netpoll.Context, error) {
		reader := bufio.NewReader(conn)
		rw := bufio.NewReadWriter(reader, bufio.NewWriter(conn))
		return &Context{reader: reader, conn: conn, rw: rw}, nil
	})
	h.SetServe(func(context netpoll.Context) error {
		ctx := context.(*Context)
		ctx.serving.Lock()
		req, err := http.ReadRequest(ctx.reader)
		if err != nil {
			ctx.serving.Unlock()
			return err
		}
		res := mux.NewResponse(req, ctx.conn, ctx.rw)
		handler.ServeHTTP(res, req)
		res.FinishRequest()
		ctx.serving.Unlock()
		mux.FreeResponse(res)
		return nil
	})
	return netpoll.ListenAndServe("tcp", addr, h)
}