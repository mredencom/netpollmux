package inproc

import (
	"net"
	"sync"
	"testing"
	"time"
)

func TestINPROC(t *testing.T) {
	address := ":9999"
	if _, err := Dial(address); err == nil {
		t.Error(err)
	}
	l, err := Listen(address)
	if err != nil {
		t.Error(err)
	}
	if _, err := Listen(address); err == nil {
		t.Error()
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			wg.Add(1)
			go func(conn net.Conn) {
				defer wg.Done()
				buf := make([]byte, 1024)
				for {
					n, err := conn.Read(buf)
					if err != nil {
						break
					}
					conn.Write(buf[:n])
				}
				conn.Close()
			}(conn)
		}
	}()
	conn, err := Dial(address)
	if err != nil {
		t.Error(err)
	}
	conn.SetWriteDeadline(time.Now().Add(time.Second))
	conn.SetReadDeadline(time.Now().Add(time.Second))
	conn.SetDeadline(time.Now().Add(time.Second))
	conn.LocalAddr()
	raddr := conn.RemoteAddr()
	raddr.Network()
	raddr.String()

	str := "Hello World"
	conn.Write([]byte(str))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		t.Error(err)
	}
	if string(buf[:n]) != str {
		t.Errorf("error %s != %s", string(buf[:n]), str)
	}
	time.Sleep(time.Millisecond)
	conn.Close()
	l.Addr()
	l.Close()
	wg.Wait()
}
