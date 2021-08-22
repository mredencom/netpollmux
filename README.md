# netpollmux
[![PkgGoDev](https://pkg.go.dev/badge/github.com/php2go/netpollmux)](https://pkg.go.dev/github.com/php2go/netpollmux)
[![Build Status](https://github.com/php2go/netpollmux/workflows/build/badge.svg)](https://github.com/php2go/netpollmux/actions)
[![codecov](https://codecov.io/gh/php2go/netpollmux/branch/master/graph/badge.svg)](https://codecov.io/gh/php2go/netpollmux)
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

#### [TLS](http://github.com/hslam/socket "socket") Example
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


#### [HTTP](http://github.com/hslam/rum "rum") Example
```go
package main

import (
	"bufio"
	"github.com/php2go/netpollmux/internal/response"
	"github.com/php2go/netpollmux/mux"
	"github.com/php2go/netpollmux/netpoll"
	"net"
	"net/http"
	"sync"
)

func main() {
	m := mux.New()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})
	ListenAndServe(":8080", m)
}

func ListenAndServe(addr string, handler http.Handler) error {
	var h = &netpoll.ConnHandler{}
	type Context struct {
		reader  *bufio.Reader
		rw      *bufio.ReadWriter
		conn    net.Conn
		serving sync.Mutex
	}
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
		res := response.NewResponse(req, ctx.conn, ctx.rw)
		handler.ServeHTTP(res, req)
		res.FinishRequest()
		ctx.serving.Unlock()
		response.FreeResponse(res)
		return nil
	})
	return netpoll.ListenAndServe("tcp", addr, h)
}
```


