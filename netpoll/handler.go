package netpoll

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

// ErrHandlerFunc is the error when the HandlerFunc is nil
var ErrHandlerFunc = errors.New("HandlerFunc must be not nil")

// ErrUpgradeFunc is the error when the Upgrade func is nil
var ErrUpgradeFunc = errors.New("Upgrade function must be not nil")

// ErrServeFunc is the error when the Serve func is nil

var ErrServeFunc = errors.New("Serve function must be not nil")

var (
	buffers = sync.Map{}
	assign  int32
)

func assignPool(size int) *sync.Pool {
	for {
		if p, ok := buffers.Load(size); ok {
			return p.(*sync.Pool)
		}
		if atomic.CompareAndSwapInt32(&assign, 0, 1) {
			var pool = &sync.Pool{New: func() interface{} {
				return make([]byte, size)
			}}
			buffers.Store(size, pool)
			atomic.StoreInt32(&assign, 0)
			return pool
		}
	}
}

// Context is returned by Upgrade for serving.
type Context interface{}

// Handler responds to a single request.
type Handler interface {
	// Upgrade upgrades the net.Conn to a Context.
	Upgrade(net.Conn) (Context, error)
	// Serve should serve a single request with the Context.
	Serve(Context) error
}

// NewHandler returns a new Handler.
func NewHandler(upgrade func(net.Conn) (Context, error), serve func(Context) error) Handler {
	return &ConnHandler{upgrade: upgrade, serve: serve}
}

// ConnHandler implements the Handler interface.
type ConnHandler struct {
	upgrade func(net.Conn) (Context, error)
	serve   func(Context) error
}

func NewConHandler() *ConnHandler {
	return &ConnHandler{}
}

// SetUpgrade sets the Upgrade function for upgrading the net.Conn.
func (h *ConnHandler) SetUpgrade(upgrade func(conn net.Conn) (Context, error)) *ConnHandler {
	h.upgrade = upgrade
	return h
}

// SetServe sets the Serve function for once serving.
func (h *ConnHandler) SetServe(serve func(Context) error) *ConnHandler {
	h.serve = serve
	return h
}

// Upgrade implements the Handler Upgrade method.
func (h *ConnHandler) Upgrade(conn net.Conn) (Context, error) {
	if h.upgrade == nil {
		return nil, ErrUpgradeFunc
	}
	return h.upgrade(conn)
}

// Serve implements the Handler Serve method.
func (h *ConnHandler) Serve(ctx Context) error {
	if h.serve == nil {
		return ErrServeFunc
	}
	return h.serve(ctx)
}

// DataHandler implements the Handler interface.
type DataHandler struct {
	// NoShared disables the DataHandler to use the buffer pool for high performance.
	// Default NoShared is false to use the buffer pool for low memory usage.
	NoShared bool
	// NoCopy returns the bytes underlying buffer when NoCopy is true,
	// The bytes returned is shared by all invocations of Read, so do not modify it.
	// Default NoCopy is false to make a copy of data for every invocations of Read.
	NoCopy bool
	// BufferSize represents the buffer size.
	BufferSize int
	upgrade    func(net.Conn) (net.Conn, error)
	// HandlerFunc is the data Serve function.
	HandlerFunc func(req []byte) (res []byte)
}

type context struct {
	reading sync.Mutex
	writing sync.Mutex
	upgrade bool
	conn    net.Conn
	pool    *sync.Pool
	buffer  []byte
}

// SetUpgrade sets the Upgrade function for upgrading the net.Conn.
func (h *DataHandler) SetUpgrade(upgrade func(net.Conn) (net.Conn, error)) {
	h.upgrade = upgrade
}

// Upgrade sets the net.Conn to a Context.
func (h *DataHandler) Upgrade(conn net.Conn) (Context, error) {
	if h.BufferSize < 1 {
		h.BufferSize = bufferSize
	}
	if h.HandlerFunc == nil {
		return nil, ErrHandlerFunc
	}
	var upgrade bool
	if h.upgrade != nil {
		c, err := h.upgrade(conn)
		if err != nil {
			return nil, err
		} else if c != nil && c != conn {
			upgrade = true
			conn = c
		}
	}
	var ctx = &context{upgrade: upgrade, conn: conn}
	if h.NoShared {
		ctx.buffer = make([]byte, h.BufferSize)
	} else {
		ctx.pool = assignPool(h.BufferSize)
	}
	return ctx, nil
}

// Serve should serve a single request with the Context ctx.
func (h *DataHandler) Serve(ctx Context) error {
	c := ctx.(*context)
	var conn = c.conn
	var n int
	var err error
	var buf []byte
	var req []byte
	if h.NoShared {
		buf = c.buffer
	} else {
		buf = c.pool.Get().([]byte)
	}
	if c.upgrade {
		c.reading.Lock()
	}
	n, err = conn.Read(buf)
	if c.upgrade {
		c.reading.Unlock()
	}
	if err != nil {
		if !h.NoShared {
			c.pool.Put(buf)
		}
		return err
	}
	req = buf[:n]
	if !h.NoCopy {
		req = make([]byte, n)
		copy(req, buf[:n])
	}
	res := h.HandlerFunc(req)
	if c.upgrade {
		c.writing.Lock()
	}
	_, err = conn.Write(res)
	if c.upgrade {
		c.writing.Unlock()
	}
	if !h.NoShared {
		c.pool.Put(buf)
	}
	return err
}
