package websocket

import (
	"net"
	"net/http"
	"reflect"
	"sync"
	"testing"
)

func TestFrame(t *testing.T) {
	{
		f := frame{FIN: 1, Opcode: BinaryFrame, PayloadData: make([]byte, 64)}
		data, _ := f.Marshal(nil)
		var f2 = frame{}
		f2.Unmarshal(data)
		if f.FIN != f2.FIN {
			t.Error()
		}
		if f.Opcode != f2.Opcode {
			t.Error()
		}
		if !reflect.DeepEqual(f.PayloadData, f2.PayloadData) {
			t.Error()
		}
		for i := 0; i < len(data); i++ {
			var f3 = frame{}
			n, _ := f3.Unmarshal(data[:i])
			if n != 0 {
				t.Error(n, len(data))
			}
		}
	}
	{
		buf := make([]byte, 64*1024)
		f := frame{FIN: 1, Opcode: BinaryFrame, Mask: 1, MaskingKey: []byte{1, 2, 3, 4}, PayloadData: make([]byte, 64)}
		cp := frame{FIN: 1, Opcode: BinaryFrame, Mask: 1, MaskingKey: []byte{1, 2, 3, 4}, PayloadData: make([]byte, 64)}
		data, _ := cp.Marshal(buf)
		var f2 = frame{}
		f2.Unmarshal(data)
		if f.FIN != f2.FIN {
			t.Error()
		}
		if f.Opcode != f2.Opcode {
			t.Error()
		}
		if f.Mask != f2.Mask {
			t.Error()
		}
		if !reflect.DeepEqual(f.MaskingKey, f2.MaskingKey) {
			t.Error()
		}
		if !reflect.DeepEqual(f.PayloadData, f2.PayloadData) {
			t.Error()
		}
		for i := 0; i < len(data); i++ {
			var f3 = frame{}
			n, _ := f3.Unmarshal(data[:i])
			if n != 0 {
				t.Error(n, len(data))
			}
		}
	}
	{
		buf := make([]byte, 64*1024)
		f := frame{FIN: 1, Opcode: BinaryFrame, PayloadData: make([]byte, 512)}
		data, _ := f.Marshal(buf)
		var f2 = frame{}
		f2.Unmarshal(data)
		if f.FIN != f2.FIN {
			t.Error()
		}
		if f.Opcode != f2.Opcode {
			t.Error()
		}
		if !reflect.DeepEqual(f.PayloadData, f2.PayloadData) {
			t.Error()
		}
		for i := 0; i < len(data); i++ {
			var f3 = frame{}
			n, _ := f3.Unmarshal(data[:i])
			if n != 0 {
				t.Error(n, len(data))
			}
		}
	}
	{
		buf := make([]byte, 128*1024)
		f := frame{FIN: 1, Opcode: BinaryFrame, PayloadData: make([]byte, 64*1024)}
		data, _ := f.Marshal(buf)
		var f2 = frame{}
		f2.Unmarshal(data)
		if f.FIN != f2.FIN {
			t.Error()
		}
		if f.Opcode != f2.Opcode {
			t.Error()
		}
		if !reflect.DeepEqual(f.PayloadData, f2.PayloadData) {
			t.Error()
		}
		for i := 0; i < len(data); i++ {
			var f3 = frame{}
			n, _ := f3.Unmarshal(data[:i])
			if n != 0 {
				t.Error(n, len(data))
			}
		}
	}
	{
		buf := make([]byte, 128*1024)
		f := frame{FIN: 1, Opcode: BinaryFrame, Mask: 1, MaskingKey: []byte{1, 2, 3, 4}, PayloadData: make([]byte, 64*1024)}
		cp := frame{FIN: 1, Opcode: BinaryFrame, Mask: 1, MaskingKey: []byte{1, 2, 3, 4}, PayloadData: make([]byte, 64*1024)}
		data, _ := cp.Marshal(buf)
		var f2 = frame{}
		f2.Unmarshal(data)
		if f.FIN != f2.FIN {
			t.Error()
		}
		if f.Opcode != f2.Opcode {
			t.Error()
		}
		if f.Mask != f2.Mask {
			t.Error()
		}
		if !reflect.DeepEqual(f.MaskingKey, f2.MaskingKey) {
			t.Error()
		}
		if !reflect.DeepEqual(f.PayloadData, f2.PayloadData) {
			t.Error()
		}
		for i := 0; i < len(data); i++ {
			var f3 = frame{}
			n, _ := f3.Unmarshal(data[:i])
			if n != 0 {
				t.Error(n, len(data))
			}
		}
	}
}

func TestConnFrame(t *testing.T) {
	network := "tcp"
	addr := ":8080"
	Serve := func(conn *Conn) {
		for {
			msg, err := conn.ReadMessage(nil)
			if err != nil {
				break
			}
			conn.WriteMessage(msg)
		}
		conn.Close()
	}

	httpServer := &http.Server{
		Addr:    addr,
		Handler: Handler(Serve),
	}
	l, _ := net.Listen(network, addr)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		httpServer.Serve(l)
	}()
	{
		conn, err := Dial(network, addr, "/", nil)
		if err != nil {
			t.Error(err)
		}
		sizes := []int{64, 512, 64 * 1024}
		for i := 0; i < len(sizes); i++ {
			msg := string(make([]byte, sizes[i]))
			if err := conn.WriteMessage([]byte(msg)); err != nil {
				t.Error(err)
			}
			buf := make([]byte, 512)
			data, err := conn.ReadMessage(buf)
			if err != nil {
				t.Error(err)
			} else if string(data) != msg {
				t.Error(string(data))
			}
		}
		conn.Close()
	}
	{
		conn, err := Dial(network, addr, "/", nil)
		if err != nil {
			t.Error(err)
		}
		conn.Close()
		msg := "Hello World"
		if err := conn.WriteMessage([]byte(msg)); err == nil {
			t.Error()
		}
	}
	{
		conn, err := Dial(network, addr, "/", nil)
		if err != nil {
			t.Error(err)
		}
		msg := "Hello World"
		if err := conn.WriteMessage([]byte(msg)); err != nil {
			t.Error(err)
		}
		conn.Close()
		_, err = conn.ReadMessage(nil)
		if err == nil {
			t.Error(err)
		}
	}
	httpServer.Close()
	wg.Wait()
}

func BenchmarkFrameMarshal(b *testing.B) {
	buf := make([]byte, 64*1024)
	for i := 0; i < b.N; i++ {
		f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: make([]byte, 512)}
		f.Marshal(buf)
	}
}

func BenchmarkFrameUnmarshal(b *testing.B) {
	buf := make([]byte, 64*1024)
	f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: make([]byte, 512)}
	data, _ := f.Marshal(buf)
	for i := 0; i < b.N; i++ {
		var f2 = &frame{}
		f2.Unmarshal(data)
	}
}

func BenchmarkFrame(b *testing.B) {
	buf := make([]byte, 64*1024)
	f := &frame{FIN: 1, Opcode: BinaryFrame, PayloadData: make([]byte, 512)}
	for i := 0; i < b.N; i++ {
		data, _ := f.Marshal(buf)
		var f2 = &frame{}
		f2.Unmarshal(data)
	}
}
