package writer

import (
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNoConcurrency(t *testing.T) {
	r, w := io.Pipe()
	size := 0
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 65536)
		for {
			n, err := r.Read(buf)
			if err != nil {
				break
			}
			size += n
		}
		close(done)
	}()
	writer := NewWriter(w, nil, 0, false)
	msg := make([]byte, 512)
	wg := sync.WaitGroup{}
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				writer.Write(msg)
			}
		}()
	}
	wg.Wait()
	writer.Close()
	w.Close()
	<-done
	if size != 512*100*64 {
		t.Error(size)
	}
}

func TestConcurrency(t *testing.T) {
	testConcurrency(1, t)
	testConcurrency(4, t)
	testConcurrency(16, t)
	testConcurrency(32, t)
}

func testConcurrency(batch int, t *testing.T) {
	r, w := io.Pipe()
	count := int64(0)
	concurrency := func() int {
		return int(atomic.LoadInt64(&count))
	}
	size := 0
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 65536)
		for {
			n, err := r.Read(buf)
			if err != nil {
				break
			}
			size += n
		}
		close(done)
	}()
	writer := NewWriter(w, concurrency, 0, false)
	msg := make([]byte, 512)
	wg := sync.WaitGroup{}
	for i := 0; i < batch; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				atomic.AddInt64(&count, 1)
				writer.Write(msg)
				atomic.AddInt64(&count, -1)
			}
		}()
	}
	wg.Wait()
	writer.Close()
	w.Close()
	n, err := writer.Write(msg)
	if n != 0 || err == nil {
		t.Error(n, err)
	}
	<-done
	if size != 512*100*batch {
		t.Error(size)
	}
}

func TestRun(t *testing.T) {
	r, w := io.Pipe()
	count := int64(0)
	concurrency := func() int {
		return int(atomic.LoadInt64(&count))
	}
	size := 0
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 65536)
		for {
			n, err := r.Read(buf)
			if err != nil {
				break
			}
			size += n
		}
		close(done)
	}()
	writer := NewWriter(w, concurrency, 0, false)
	msg := make([]byte, 512)
	wg := sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				atomic.AddInt64(&count, 1)
				writer.Write(msg)
				atomic.AddInt64(&count, -1)
				time.Sleep(time.Millisecond * 101)
			}
		}()
	}
	wg.Wait()
	writer.Close()
	w.Close()
	<-done
	if size != 512*10*1 {
		t.Error(size)
	}
}

func TestShared(t *testing.T) {
	r, w := io.Pipe()
	count := int64(0)
	concurrency := func() int {
		return int(atomic.LoadInt64(&count))
	}
	size := 0
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 65536)
		for {
			n, err := r.Read(buf)
			if err != nil {
				break
			}
			size += n
		}
		close(done)
	}()
	writer := NewWriter(w, concurrency, 65536, true)
	msg := make([]byte, 512)
	wg := sync.WaitGroup{}
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				atomic.AddInt64(&count, 1)
				writer.Write(msg)
				atomic.AddInt64(&count, -1)
			}
		}()
	}
	wg.Wait()
	writer.Close()
	w.Close()
	<-done
	if size != 512*100*64 {
		t.Error(size)
	}
}

func TestMMS(t *testing.T) {
	testSize(1, t)
	testSize(512+1, t)
	testSize(512*4+1, t)
	testSize(512*16+1, t)
	testSize(512*28+1, t)
	testSize(512*32+1, t)
}

func testSize(mms int, t *testing.T) {
	r, w := io.Pipe()
	count := int64(0)
	concurrency := func() int {
		return int(atomic.LoadInt64(&count))
	}
	size := 0
	done := make(chan struct{})
	go func() {
		buf := make([]byte, 65536)
		for {
			n, err := r.Read(buf)
			if err != nil {
				break
			}
			size += n
		}
		close(done)
	}()
	writer := NewWriter(w, concurrency, mms, true)
	msg := make([]byte, 512)
	wg := sync.WaitGroup{}
	for i := 0; i < 64; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				atomic.AddInt64(&count, 1)
				writer.Write(msg)
				atomic.AddInt64(&count, -1)
			}
		}()
	}
	wg.Wait()
	writer.Close()
	writer.Close()
	w.Close()
	<-done
	if size != 512*100*64 {
		t.Error(size)
	}
}
