package response

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"
)

func testHTTP(method, url string, status int, result string, t *testing.T) {
	var req *http.Request
	var err error
	req, err = http.NewRequest(method, url, nil)
	if err != nil {
		t.Error(err)
	}
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

func testHeader(method, url string, status int, result string, header map[string]string, t *testing.T) {
	var req *http.Request
	var err error
	req, err = http.NewRequest(method, url, nil)
	if err != nil {
		t.Error(err)
	}
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
	} else {
		for k, v := range header {
			value := resp.Header.Get(k)
			if value != v {
				t.Error(k, v, value)
			}
		}
	}
}

func testMultipart(url string, status int, result string, values map[string]io.Reader, t *testing.T) {
	var b bytes.Buffer
	var err error
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				t.Error(err)
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				t.Error(err)
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			t.Error(err)
		}
	}
	w.Close()
	var req *http.Request
	req, err = http.NewRequest("POST", url, &b)
	if err != nil {
		t.Error(err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			MaxConnsPerHost:   1,
			DisableKeepAlives: true,
		},
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
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

var testConnected = "200 Connected to Server"

func TestResponse(t *testing.T) {
	m := http.NewServeMux()
	length := 1024 * 64
	var msg = make([]byte, length)
	for i := 0; i < length; i++ {
		msg[i] = 'a'
	}
	m.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!\r\n"))
	})
	m.HandleFunc("/chunked", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Write([]byte("Hello"))
		w.Write([]byte(" World!\r\n"))
	})
	m.HandleFunc("/msg", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentLength, strconv.FormatInt(int64(len(msg)), 10))
		w.Header().Set(contentType, defaultContentType)
		if n, err := w.Write(msg); err != nil {
			t.Error(err)
		} else if n != len(msg) {
			t.Errorf("length error %d %d", n, len(msg))
		}
		if n, err := w.Write(nil); err != nil {
			t.Error(err)
		} else if n != 0 {
			t.Errorf("length error %d %d", n, 0)
		}
	})
	m.HandleFunc("/length", func(w http.ResponseWriter, r *http.Request) {
		length := len(msg) / 2
		w.Header().Set(contentLength, strconv.FormatInt(int64(length), 10))
		if _, err := w.Write(msg); err == nil {
			t.Error()
		} else {
			if n, err := w.Write(msg[:length]); err != nil {
				t.Error(err)
			} else if n != length {
				t.Errorf("length error %d %d", n, len(msg))
			}
		}
	})
	m.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentLength, "invalid content length header")
		w.WriteHeader(http.StatusBadRequest)
		w.WriteHeader(http.StatusOK)
	})
	m.HandleFunc("/connect", func(w http.ResponseWriter, r *http.Request) {
		testConnect(w, r, t)
		if _, _, err := w.(http.Hijacker).Hijack(); err != http.ErrHijacked {
			t.Error(err)
		}
		_, err := w.Write(nil)
		if err != http.ErrHijacked {
			t.Error(err)
		}
		w.WriteHeader(0)
	})
	m.HandleFunc("/header", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	})
	m.HandleFunc("/multipart", func(w http.ResponseWriter, r *http.Request) {
		r.ParseMultipartForm(1024)
		mf := r.MultipartForm
		if mf != nil {
			w.Write([]byte(mf.Value["value"][0]))
		}
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
				reader := NewBufioReader(conn)
				writer := NewBufioWriter(conn)
				var err error
				var req *http.Request
				for err == nil {
					req, err = http.ReadRequest(reader)
					if err != nil {
						break
					}
					res := NewResponse(req, conn, bufio.NewReadWriter(reader, writer))
					m.ServeHTTP(res, req)
					res.FinishRequest()
					res.FinishRequest()
					FreeResponse(res)
				}
				FreeBufioReader(reader)
				FreeBufioWriter(writer)
			}(conn)
		}
	}()
	time.Sleep(time.Millisecond * 10)
	testHTTP("GET", "http://"+addr+"/", http.StatusOK, "Hello World!\r\n", t)
	testHTTP("GET", "http://"+addr+"/chunked", http.StatusOK, "Hello World!\r\n", t)
	testHTTP("GET", "http://"+addr+"/msg", http.StatusOK, string(msg), t)
	testHTTP("GET", "http://"+addr+"/length", http.StatusOK, string(msg[:len(msg)/2]), t)
	testHTTP("GET", "http://"+addr+"/error", http.StatusBadRequest, "", t)
	testDialHTTPPath(addr, "/connect", t)
	header := make(map[string]string)
	header["Access-Control-Allow-Origin"] = "*"
	testHeader("GET", "http://"+addr+"/header", http.StatusOK, "", header, t)
	values := make(map[string]io.Reader)
	values["value"] = bytes.NewReader(msg)
	testMultipart("http://"+addr+"/multipart", http.StatusOK, string(msg), values, t)
	ln.Close()
	wg.Wait()
}

func testConnect(w http.ResponseWriter, r *http.Request, t *testing.T) {
	if r.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		t.Error("405 must CONNECT: ", r.Method)
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		t.Error(err)
	}
	io.WriteString(conn, "HTTP/1.0 "+testConnected+"\n\n")
	_, err = conn.Write([]byte("PING"))
	if err != nil {
		t.Error(err)
	}
	buf := make([]byte, 1024)
	if n, err := conn.Read(buf); err != nil {
		t.Error(err)
	} else if string(buf[:n]) != "PONG" {
		t.Error(string(buf[:n]))
	}
	conn.Close()
}

func testDialHTTPPath(address, path string, t *testing.T) {
	var err error
	var network = "tcp"
	conn, err := net.Dial(network, address)
	if err != nil {
		t.Error(err)
	}
	io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")
	reader := bufio.NewReader(conn)
	// Require successful HTTP response
	// before switching to Your protocol.
	resp, err := http.ReadResponse(reader, &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == testConnected {
		buf := make([]byte, 1024)
		if n, err := reader.Read(buf); err != nil {
			t.Error(err)
		} else if string(buf[:n]) != "PING" {
			t.Error(string(buf[:n]))
		}
		conn.Write([]byte("PONG"))
		return
	}
	if err == nil {
		t.Error("unexpected HTTP response: " + resp.Status)
	}
	conn.Close()
}

func TestNewBufioReader(t *testing.T) {
	reader := NewBufioReader(nil)
	FreeBufioReader(reader)
}

func TestFreeResponse(t *testing.T) {
	FreeResponse(nil)
}

func TestFreeHeader(t *testing.T) {
	freeHeader(nil)
	h := make(http.Header)
	h.Add("Content-Type", defaultContentType)
	freeHeader(h)
}

func TestBodyAllowed(t *testing.T) {
	defer func() {
		e := recover()
		if e == nil {
			t.Error("should panic")
		}
	}()
	res := &Response{}
	res.bodyAllowed()
}

func TestBodyAllowedForStatus(t *testing.T) {
	if bodyAllowedForStatus(100) {
		t.Error(100)
	}
	if bodyAllowedForStatus(204) {
		t.Error(204)
	}
	if bodyAllowedForStatus(304) {
		t.Error(304)
	}
}

func TestCheckWriteHeaderCode(t *testing.T) {
	defer func() {
		e := recover()
		if e == nil {
			t.Error("should panic")
		}
	}()
	checkWriteHeaderCode(0)
}
