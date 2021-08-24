package netpoll_benchmark

import (
	"flag"
	"time"

	"github.com/php2go/netpollmux/netpoll"
)

var addr string
var async bool
var sleep int

func init() {
	flag.StringVar(&addr, "addr", ":9999", "-addr=:9999")
	flag.BoolVar(&async, "async", false, "-async=false")
	flag.IntVar(&sleep, "sleep", 0, "-sleep=0")
	flag.Parse()
}

func main() {
	var handler = &netpoll.DataHandler{
		NoShared:   true,
		NoCopy:     true,
		BufferSize: 1024,
		HandlerFunc: func(req []byte) (res []byte) {
			if sleep > 0 {
				time.Sleep(time.Millisecond * time.Duration(sleep))
			}
			res = req
			return
		},
	}
	if err := netpoll.ListenAndServe("tcp", addr, handler); err != nil {
		panic(err)
	}
}
