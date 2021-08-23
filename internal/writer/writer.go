package writer

import (
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/php2go/netpollmux/internal/buffer"
)

const thresh = 4
const maximumSegmentSize = 65536
const lastsSize = 4

// Flusher is the interface that wraps the basic Flush method.
//
// Flush writes any buffered data to the underlying io.Writer.
type Flusher interface {
	Flush() (err error)
}

// Writer implements batch writing for an io.Writer object.
type Writer struct {
	shared      bool
	mu          *sync.Mutex
	writer      io.Writer
	mss         int
	buffer      []byte
	size        int
	count       int
	writeCnt    int
	concurrency func() int
	thresh      int
	lasts       [lastsSize]int
	cursor      int
	trigger     chan struct{}
	done        chan struct{}
	closed      int32
}

// NewWriter returns a new batch Writer with the concurrency.
func NewWriter(writer io.Writer, concurrency func() int, size int, shared bool) *Writer {
	if size < 1 {
		size = maximumSegmentSize
	}
	w := &Writer{writer: writer}
	if concurrency != nil {
		var buffer []byte
		if !shared {
			buffer = make([]byte, size)
		}
		w.shared = shared
		w.mu = &sync.Mutex{}
		w.thresh = thresh
		w.mss = size
		w.buffer = buffer
		w.concurrency = concurrency
		w.trigger = make(chan struct{})
		w.done = make(chan struct{})
		go w.run()
	}
	return w
}

func (w *Writer) batch() (n int) {
	w.cursor++
	w.lasts[w.cursor%lastsSize] = w.concurrency()
	var max int
	for i := 0; i < lastsSize; i++ {
		if w.lasts[i] > max {
			max = w.lasts[i]
		}
	}
	return max
}

// Write writes the contents of p into the buffer or the underlying io.Writer.
// It returns the number of bytes written.
func (w *Writer) Write(p []byte) (n int, err error) {
	if w.concurrency == nil {
		return w.writer.Write(p)
	}
	batch := w.batch()
	length := len(p)
	w.mu.Lock()
	defer w.mu.Unlock()
	w.writeCnt++
	if w.size+length > w.mss {
		if w.size > 0 {
			w.flush(false)
		}
		if length > 0 {
			_, err = w.writer.Write(p)
		}
	} else if batch <= w.thresh {
		if w.size > 0 {
			copy(w.buffer[w.size:], p)
			w.size += length
			err = w.flush(false)
		} else {
			n, err = w.writer.Write(p)
			w.size = 0
			w.count = 0
		}
	} else if batch <= w.thresh*w.thresh {
		if w.writeCnt < w.thresh {
			if w.size > 0 {
				copy(w.buffer[w.size:], p)
				w.size += length
				err = w.flush(false)
			} else {
				_, err = w.writer.Write(p)
				w.size = 0
				w.count = 0
			}
		} else {
			if w.shared && len(w.buffer) == 0 {
				w.buffer = buffer.AssignPool(w.mss).GetBuffer()
			}
			copy(w.buffer[w.size:], p)
			w.size += length
			w.count++
			if w.count > batch-w.thresh {
				err = w.flush(true)
			}
			select {
			case w.trigger <- struct{}{}:
			default:
			}
		}
	} else {
		alpha := w.thresh*2 - (batch-1)/w.thresh
		if alpha > 1 {
			if w.writeCnt < alpha {
				if w.size > 0 {
					copy(w.buffer[w.size:], p)
					w.size += length
					err = w.flush(false)
				} else {
					_, err = w.writer.Write(p)
					w.size = 0
					w.count = 0
				}
			} else {
				if w.shared && len(w.buffer) == 0 {
					w.buffer = buffer.AssignPool(w.mss).GetBuffer()
				}
				copy(w.buffer[w.size:], p)
				w.size += length
				w.count++
				if w.count > batch-alpha {
					err = w.flush(true)
				}
				select {
				case w.trigger <- struct{}{}:
				default:
				}
			}
		} else {
			if w.shared && len(w.buffer) == 0 {
				w.buffer = buffer.AssignPool(w.mss).GetBuffer()
			}
			copy(w.buffer[w.size:], p)
			w.size += length
			w.count++
			if w.count > batch-1 {
				err = w.flush(true)
			}
			select {
			case w.trigger <- struct{}{}:
			default:
			}
		}
	}
	if err != nil {
		return 0, err
	}
	return len(p), err
}

// Flush writes any buffered data to the underlying io.Writer.
func (w *Writer) Flush() error {
	w.mu.Lock()
	err := w.flush(true)
	w.mu.Unlock()
	return err
}

func (w *Writer) flush(reset bool) (err error) {
	if w.size > 0 {
		_, err = w.writer.Write(w.buffer[:w.size])
		if w.shared {
			buffer.AssignPool(w.mss).PutBuffer(w.buffer)
			w.buffer = nil
		}
		w.size = 0
		w.count = 0
		if reset {
			w.writeCnt = 0
		}
	}
	return
}

func (w *Writer) run() {
	for {
		w.mu.Lock()
		w.flush(true)
		w.mu.Unlock()
		var d time.Duration
		if w.batch() < w.thresh*2 {
			d = time.Second
		} else {
			d = time.Microsecond * 100
		}
		timer := time.NewTimer(d)
		select {
		case <-timer.C:
		case <-w.trigger:
			timer.Stop()
			time.Sleep(time.Microsecond * time.Duration(w.batch()))
		case <-w.done:
			timer.Stop()
			return
		}
	}
}

// Close closes the writer, but do not close the underlying io.Writer
func (w *Writer) Close() (err error) {
	if w.concurrency != nil {
		w.Flush()
	}
	if !atomic.CompareAndSwapInt32(&w.closed, 0, 1) {
		return err
	}
	if w.concurrency != nil {
		close(w.done)
	}
	return err
}
