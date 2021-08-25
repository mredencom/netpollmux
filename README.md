# netpollmux
[![PkgGoDev](https://pkg.go.dev/badge/github.com/php2go/netpollmux)](https://pkg.go.dev/github.com/php2go/netpollmux)
[![Go Report Card](https://goreportcard.com/badge/github.com/php2go/netpollmux)](https://goreportcard.com/report/github.com/php2go/netpollmux)
[![LICENSE](https://img.shields.io/github/license/php2go/netpollmux.svg?style=flat-square)](https://github.com/php2go/netpollmux/blob/master/LICENSE)

Package netpollmux implements a network poller based on epoll/kqueue.

## Features

* Epoll/kqueue
* TCP/UNIX
* Compatible with the net.Conn interface.
* Upgrade connection
* Non-blocking I/O
* Splice/sendfile
* Rescheduling workers

**Comparison to other packages.**

|Package| [net](https://github.com/golang/go/tree/master/src/net "net")| [netpollmux](https://github.com/php2go/netpollmux "netpoll")|[gnet](https://github.com/panjf2000/gnet "gnet")|[evio](https://github.com/tidwall/evio "evio")|
|:--:|:--|:--|:--|:--|
|Low memory usage|No|Yes|Yes|Yes|
|Non-blocking I/O|No|Yes|Yes|Yes|
|Splice/sendfile|Yes|Yes|No|No|
|Rescheduling|Yes|Yes|No|No|
|Compatible with the net.Conn interface|Yes|Yes|No|No|

## [Benchmark](http://github.com/hslam/netpoll-benchmark "netpoll-benchmark")

<img src="https://raw.githubusercontent.com/php2go/netpollmux/main/netpoll-qps.png" width = "400" height = "300" alt="mock 0ms" align=center><img src="https://raw.githubusercontent.com/php2go/netpollmux/main/netpoll-mock-time-qps.png" width = "400" height = "300" alt="mock 1ms" align=center>

## Get started

### Install
```
go get github.com/php2go/netpollmux
```
### Import
```
import "github.com/php2go/netpollmux"
```
### Usage
#### Simple Example
```go
package main

import "github.com/php2go/netpollmux"

func main() {
	var handler = &netpoll.DataHandler{
		NoShared:   true,
		NoCopy:     true,
		BufferSize: 1024,
		HandlerFunc: func(req []byte) (res []byte) {
			res = req
			return
		},
	}
	if err := netpoll.ListenAndServe("tcp", ":9999", handler); err != nil {
		panic(err)
	}
}
```

#### TLS Example
```go
package main

import (
	"crypto/tls"
	"github.com/php2go/netpollmux/internal/socket"
	"github.com/php2go/netpollmux/netpoll"
	"net"
)

func main() {
	var handler = &netpoll.DataHandler{
		NoShared:   true,
		NoCopy:     true,
		BufferSize: 1024,
		HandlerFunc: func(req []byte) (res []byte) {
			res = req
			return
		},
	}
	handler.SetUpgrade(func(conn net.Conn) (net.Conn, error) {
		tlsConn := tls.Server(conn, socket.DefalutTLSConfig())
		if err := tlsConn.Handshake(); err != nil {
			return nil, err
		}
		return tlsConn, nil
	})
	if err := netpoll.ListenAndServe("tcp", ":9999", handler); err != nil {
		panic(err)
	}
}
```

#### Websocket Example
```go
package main

import (
	"github.com/php2go/netpollmux/internal/websocket"
	"github.com/php2go/netpollmux/netpoll"
	"net"
)

func main() {
	var handler = &netpoll.ConnHandler{}
	handler.SetUpgrade(func(conn net.Conn) (netpoll.Context, error) {
		return websocket.Upgrade(conn, nil)
	})
	handler.SetServe(func(context netpoll.Context) error {
		ws := context.(*websocket.Conn)
		msg, err := ws.ReadMessage()
		if err != nil {
			return err
		}
		return ws.WriteMessage(msg)
	})
	if err := netpoll.ListenAndServe("tcp", ":9999", handler); err != nil {
		panic(err)
	}
}
```


#### HTTP Example
```go
package main

import (
	"bufio"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/julienschmidt/httprouter"
	"github.com/php2go/netpollmux/internal/logger"
	"github.com/php2go/netpollmux/mux"
	"github.com/php2go/netpollmux/netpoll"
)

func main() {
	m := mux.NewRoute()
	m.GET("/hello/:id", func(w http.ResponseWriter, req *http.Request, params httprouter.Params) {
		pp := req.URL.Query()
		logger.Info("query params：", params, pp)
		mux.JSON(w, req, []string{"hello world"}, http.StatusOK)
	})
	log.Fatal(m.Run(":8080"))
}
```


