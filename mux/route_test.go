package mux

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"testing"
)

func TestContextKey(t *testing.T) {
	if RecoveryContextKey.String() != fmt.Sprint(RecoveryContextKey) {
		t.Error()
	}
}

func TestParseMatch(t *testing.T) {
	pattern := "/db/:key/meng/:value/huang"
	i := strings.Index(pattern, ":")
	prefix := pattern[:i]
	match := strings.Split(pattern[i:], "/")
	params := make(map[string]string)
	key := ""
	for i := 0; i < len(match); i++ {
		if strings.Contains(match[i], ":") {
			match[i] = strings.Trim(match[i], ":")
			params[match[i]] = ""
			if i > 0 {
				key += "/"
			}
		} else {
			key += "/" + match[i]
			match[i] = ""
		}
	}
	path := "/db/123/meng/456/huang"
	strs := strings.Split(strings.Trim(path, prefix), "/")
	if len(strs) == len(match) {
		for i := 0; i < len(strs); i++ {
			if match[i] != "" {
				if _, ok := params[match[i]]; ok {
					params[match[i]] = strs[i]
				}
			}
		}
	}
	if params["key"] != "123" || params["value"] != "456" {
		t.Error(params)
	}
}

func testHTTPs(method, url string, status int, result string, t *testing.T) {
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

func TestMux(t *testing.T) {
	m := NewRoute()
	m.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found : "+r.URL.String(), http.StatusNotFound)
	})
	m.Use(func(w http.ResponseWriter, r *http.Request) {
	})
	m.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(fmt.Sprintf("hello world Method:%s\n", r.Method)))
	}).All()
	m.HandleFunc("/hello/:key/world/:value", func(w http.ResponseWriter, r *http.Request) {
		params := m.Params(r)
		w.Write([]byte(fmt.Sprintf("hello key:%s value:%s\n", params["key"], params["value"])))
	}).GET().POST().PUT().DELETE()
	m.Group("/group", func(m *Route) {
		m.HandleFunc("/foo/:id", func(w http.ResponseWriter, r *http.Request) {
			params := m.Params(r)
			w.Write([]byte(fmt.Sprintf("group/foo id:%s\n", params["id"])))
		}).GET()
		m.HandleFunc("/bar/:id", func(w http.ResponseWriter, r *http.Request) {
			params := m.Params(r)
			w.Write([]byte(fmt.Sprintf("group/bar id:%s\n", params["id"])))
		}).GET()
	})
	addr := ":8080"
	httpServer := &http.Server{
		Addr:    addr,
		Handler: m,
	}
	l, _ := net.Listen("tcp", addr)
	go httpServer.Serve(l)
	testHTTP("GET", "http://"+addr+"/favicon.ico", http.StatusNotFound, "Not Found : /favicon.ico\n", t)
	testHTTP("GET", "http://"+addr+"/hello", http.StatusOK, "hello world Method:GET\n", t)
	testHTTP("POST", "http://"+addr+"/hello", http.StatusOK, "hello world Method:POST\n", t)
	testHTTP("PUT", "http://"+addr+"/hello", http.StatusOK, "hello world Method:PUT\n", t)
	testHTTP("DELETE", "http://"+addr+"/hello", http.StatusOK, "hello world Method:DELETE\n", t)
	testHTTP("PATCH", "http://"+addr+"/hello", http.StatusOK, "hello world Method:PATCH\n", t)
	testHTTP("OPTIONS", "http://"+addr+"/hello", http.StatusOK, "hello world Method:OPTIONS\n", t)
	testHTTP("TRACE", "http://"+addr+"/hello", http.StatusOK, "hello world Method:TRACE\n", t)
	testHTTP("CONNECT", "http://"+addr+"/hello", http.StatusOK, "hello world Method:CONNECT\n", t)
	if resp, err := http.Head("http://" + addr + "/hello"); err != nil {
		t.Error(err)
	} else if resp.StatusCode != http.StatusOK {
		t.Error(resp.StatusCode)
	}
	testHTTP("GET", "http://"+addr+"/hello/123/world/456", http.StatusOK, "hello key:123 value:456\n", t)
	testHTTP("GET", "http://"+addr+"/group/foo/1", http.StatusOK, "group/foo id:1\n", t)
	testHTTP("GET", "http://"+addr+"/group/bar/2", http.StatusOK, "group/bar id:2\n", t)
	httpServer.Close()
}

func TestDefaultNotFound(t *testing.T) {
	m := NewRoute()
	addr := ":8080"
	httpServer := &http.Server{
		Addr:    addr,
		Handler: m,
	}
	l, _ := net.Listen("tcp", addr)
	go httpServer.Serve(l)
	testHTTP("GET", "http://"+addr+"/favicon.ico", http.StatusNotFound, "404 Not Found : /favicon.ico\n", t)
	httpServer.Close()
}

func TestDefaultRecovery(t *testing.T) {
	m := NewRoute()
	m.Recovery(Recovery)
	msg := "panic test"
	m.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		panic(msg)
		w.Write([]byte("hello world Method:GET\n"))
	}).GET()
	addr := ":8080"
	httpServer := &http.Server{
		Addr:    addr,
		Handler: m,
	}
	l, _ := net.Listen("tcp", addr)
	go httpServer.Serve(l)
	testHTTP("GET", "http://"+addr+"/hello", http.StatusInternalServerError, "500 Internal Server Error : "+msg+"\n", t)
	httpServer.Close()
}

func TestHandleFunc(t *testing.T) {
	m := NewRoute()
	m.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world Method:GET\n"))
	}).GET()
	m.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello world Method:POST\n"))
	}).POST()
	m.HandleFunc("/hello/:key", func(w http.ResponseWriter, r *http.Request) {
		params := m.Params(r)
		w.Write([]byte(fmt.Sprintf("hello key:%s\n", params["key"])))
	}).GET()
	m.HandleFunc("/hello/:key/:value", func(w http.ResponseWriter, r *http.Request) {
		params := m.Params(r)
		w.Write([]byte(fmt.Sprintf("hello key:%s value:%s\n", params["key"], params["value"])))
	}).GET()
	m.HandleFunc("/hello/:key/world/:value", func(w http.ResponseWriter, r *http.Request) {
		params := m.Params(r)
		w.Write([]byte(fmt.Sprintf("hello world key:%s value:%s\n", params["key"], params["value"])))
	}).GET()
	m.HandleFunc("/hello/:key/world/:value/mux", func(w http.ResponseWriter, r *http.Request) {
		params := m.Params(r)
		w.Write([]byte(fmt.Sprintf("hello world mux key:%s value:%s\n", params["key"], params["value"])))
	}).GET()
	addr := ":8080"
	httpServer := &http.Server{
		Addr:    addr,
		Handler: m,
	}
	l, _ := net.Listen("tcp", addr)
	go httpServer.Serve(l)
	testHTTP("GET", "http://"+addr+"/hello", http.StatusOK, "hello world Method:GET\n", t)
	testHTTP("POST", "http://"+addr+"/hello", http.StatusOK, "hello world Method:POST\n", t)
	testHTTP("GET", "http://"+addr+"/hello/123", http.StatusOK, "hello key:123\n", t)
	testHTTP("GET", "http://"+addr+"/hello/123/456", http.StatusOK, "hello key:123 value:456\n", t)
	testHTTP("GET", "http://"+addr+"/hello/123/world/456", http.StatusOK, "hello world key:123 value:456\n", t)
	testHTTP("GET", "http://"+addr+"/hello/123/world/456/mux", http.StatusOK, "hello world mux key:123 value:456\n", t)
	httpServer.Close()
}

func TestServeHTTP(t *testing.T) {
	m := NewRoute()
	m.HandleFunc("//hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello\n"))
	})
	addr := ":8080"
	httpServer := &http.Server{
		Addr:    addr,
		Handler: m,
	}
	l, _ := net.Listen("tcp", addr)
	go httpServer.Serve(l)
	testHTTP("GET", "http://"+addr+"/hello", http.StatusOK, "hello\n", t)
	httpServer.Close()
}

func TestGroup(t *testing.T) {
	m := NewRoute()
	m.Group("/group", func(m *Route) {
		m.HandleFunc("/foo/:id", func(w http.ResponseWriter, r *http.Request) {
			params := m.Params(r)
			w.Write([]byte(fmt.Sprintf("group/foo id:%s\n", params["id"])))
		}).GET()
		m.HandleFunc("/bar/:id", func(w http.ResponseWriter, r *http.Request) {
			params := m.Params(r)
			w.Write([]byte(fmt.Sprintf("group/bar id:%s\n", params["id"])))
		}).GET()
	})
	defer func() {
		if err := recover(); err != nil {
			if err != ErrGroupExisted {
				t.Error(err)
			}
		}
	}()
	m.Group("/group", func(m *Route) {})
}

func TestParseParams(t *testing.T) {
	func() {
		m := NewRoute()
		defer func() {
			if err := recover(); err != nil {
				if err != ErrParamsKeyEmpty {
					t.Error(err)
				}
			}
		}()
		m.HandleFunc("/:", func(w http.ResponseWriter, r *http.Request) {
			params := m.Params(r)
			w.Write([]byte(fmt.Sprintf("group/foo id:%s\n", params["id"])))
		}).GET()
	}()
	func() {
		m := NewRoute()
		defer func() {
			if err := recover(); err != nil {
				if err != ErrParamsKeyEmpty {
					t.Error(err)
				}
			}
		}()
		m.HandleFunc("/:/", func(w http.ResponseWriter, r *http.Request) {
			params := m.Params(r)
			w.Write([]byte(fmt.Sprintf("group/foo id:%s\n", params["id"])))
		}).GET()
	}()
}
