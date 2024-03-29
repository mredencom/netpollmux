package splice

import (
	"errors"
	"io"
	"net"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/php2go/netpollmux/internal/buffer"
)

const (
	idleTime             = time.Second
	maxContexts          = 16384
	maxContextsPerBucket = 256

	// EAGAIN will be returned when resource temporarily unavailable.
	EAGAIN = syscall.EAGAIN
)

var (
	bucketsLen      = runtime.NumCPU()
	buckets         = make([]bucket, bucketsLen)
	maxIdleContexts = contexts(bucketsLen)

	// EOF is the error returned by Read when no more input is available.
	// Functions should return EOF only to signal a graceful end of input.
	// If the EOF occurs unexpectedly in a structured data stream,
	// the appropriate error is either ErrUnexpectedEOF or some other error
	// giving more detail.
	EOF = io.EOF

	// ErrNotHandled will be returned when the splice is not supported.
	ErrNotHandled = errors.New("The splice is not supported")
)

func contexts(bucketsLen int) int {
	if bucketsLen < maxContexts/maxContextsPerBucket {
		return maxContextsPerBucket
	}
	return maxContexts / bucketsLen
}

// MaxIdleContextsPerBucket sets the maxIdleContexts per bucket.
func MaxIdleContextsPerBucket(max int) {
	if max > 0 && max*bucketsLen < maxContexts*2 {
		maxIdleContexts = max
	}
}

// context represents a splice context.
type context struct {
	buffer []byte
	writer int
	reader int
	pool   *buffer.Pool
	bucket *bucket
	alive  bool
}

type bucket struct {
	lock            sync.Mutex
	created         bool
	pending         map[*context]struct{}
	queue           []*context
	maxIdleContexts int
	lastIdle        time.Time
}

// assignBucket assigns a bucket
func assignBucket(id int) *bucket {
	return &buckets[id%bucketsLen]
}

func (b *bucket) GetInstance() *bucket {
	b.lock.Lock()
	if !b.created {
		b.created = true
		b.pending = make(map[*context]struct{})
		b.maxIdleContexts = maxIdleContexts
		b.lock.Unlock()
		b.lastIdle = time.Now()
		go b.run()
	} else {
		b.lock.Unlock()
	}
	return b
}

func (b *bucket) run() {
	idleContexts := make([]*context, b.maxIdleContexts)
	var idles int
	for {
		if b.lastIdle.Add(idleTime).Before(time.Now()) {
			b.lock.Lock()
			if len(b.pending) == 0 {
				b.created = false
				b.lock.Unlock()
				break
			}
			b.lock.Unlock()
		}
		time.Sleep(time.Second)
		b.lock.Lock()
		idles = copy(idleContexts, b.queue[len(b.queue)/2:])
		if idles > 0 {
			b.queue = b.queue[:len(b.queue)-idles]
			for _, ctx := range idleContexts {
				delete(b.pending, ctx)
			}
		}
		b.lock.Unlock()
		if idles > 0 {
			for _, ctx := range idleContexts[:idles] {
				ctx.Close()
			}
		}
	}
}

func (b *bucket) Get() (ctx *context, err error) {
	b.lock.Lock()
	if len(b.queue) > 0 {
		ctx = b.queue[0]
		n := copy(b.queue, b.queue[1:])
		b.queue = b.queue[:n]
		b.lock.Unlock()
		b.lastIdle = time.Now()
	} else {
		b.lock.Unlock()
		ctx, err = newContext(b)
		if err == nil {
			b.lock.Lock()
			b.pending[ctx] = struct{}{}
			b.lock.Unlock()
			b.lastIdle = time.Now()
		}
	}
	return
}

func (b *bucket) Free(ctx *context) {
	b.lock.Lock()
	if len(b.queue) < b.maxIdleContexts && ctx.alive {
		b.queue = append(b.queue, ctx)
		b.lock.Unlock()
	} else {
		delete(b.pending, ctx)
		b.lock.Unlock()
		ctx.Close()
	}
}

func (b *bucket) Release() {
	b.lock.Lock()
	if b.pending == nil || len(b.pending) == 0 {
		b.lock.Unlock()
		return
	}
	pending := b.pending
	b.pending = make(map[*context]struct{})
	b.queue = []*context{}
	b.lock.Unlock()
	for ctx := range pending {
		ctx.Close()
	}
}

func spliceBuffer(dst, src net.Conn, len int64) (n int64, err error) {
	bufferSize := maxSpliceSize
	if bufferSize < int(len) {
		bufferSize = int(len)
	}
	var buf []byte
	pool := buffer.AssignPool(bufferSize)
	buf = pool.GetBuffer()
	defer pool.PutBuffer(buf)
	var remain int
	remain, err = src.Read(buf)
	if err != nil {
		return 0, err
	}
	if remain == 0 {
		return 0, EOF
	}
	var out int
	var pos int
	for remain > 0 {
		out, err = dst.Write(buf[pos : pos+remain])
		if out > 0 {
			remain -= out
			n += int64(out)
			pos += out
			continue
		}
		if err != syscall.EAGAIN {
			return n, EOF
		}
	}
	return n, nil
}
