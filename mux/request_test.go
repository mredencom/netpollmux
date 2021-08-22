package mux

import (
	"bufio"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

func testReqHTTP(method, url string, status int, result string, t *testing.T) {
	var req *http.Request
	req, _ = http.NewRequest(method, url, nil)
	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:   1,
			DisableKeepAlives: true,
		},
	}
	if resp, err := client.Do(req); err != nil {
		t.Error(err)
	} else if resp.StatusCode != status {
		t.Error(resp.StatusCode)
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Error(err)
	} else if string(body) != result {
		t.Error(string(body))
	}
}

func TestReqResponse(t *testing.T) {
	m := http.NewServeMux()
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!\r\n"))
	})
	m.HandleFunc("/chunked", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Write([]byte("Hello"))
		w.Write([]byte(" World!\r\n"))
	})
	addr := ":8080"
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		t.Error(err)
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := ln.Accept()
			if err != nil {
				break
			}
			go func(conn net.Conn) {
				reader := bufio.NewReader(conn)
				writer := bufio.NewWriter(conn)
				var err error
				var req *http.Request
				for err == nil {
					req, err = ReadFastRequest(reader)
					if err != nil {
						break
					}
					res := NewResponse(req, conn, bufio.NewReadWriter(reader, writer))
					m.ServeHTTP(res, req)
					res.FinishRequest()
					FreeRequest(req)
					FreeResponse(res)
				}
			}(conn)
		}
	}()
	time.Sleep(time.Millisecond * 10)
	testHTTP("GET", "http://"+addr+"/", http.StatusOK, "Hello World!\r\n", t)
	testHTTP("GET", "http://"+addr+"/chunked", http.StatusOK, "Hello World!\r\n", t)
	ln.Close()
	wg.Wait()
}

func TestReqFreeRequest(t *testing.T) {
	FreeRequest(nil)
}

func TestReqFreeHeader(t *testing.T) {
	freeHeader(nil)
	h := make(http.Header)
	h.Add("Content-Type", "text/plain; charset=utf-8")
	freeHeader(h)
}
