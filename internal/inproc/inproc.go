package inproc

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// network represents name of the network.
const network = "inproc"

// addr represents a network end point address.
//
// addr implements the net.Addr interface.
type addr struct {
	network string
	address string
}

// Network returns name of the network.
func (a addr) Network() string {
	return a.network
}

// String returns string form of address.
func (a addr) String() string {
	return a.address
}

// addrs contains in-process listeners.
var addrs struct {
	locker    sync.RWMutex
	listeners map[addr]*listener
}

func init() {
	addrs.listeners = make(map[addr]*listener)
}

// conn implements the net.Conn interface.
type conn struct {
	r     io.ReadCloser
	w     io.WriteCloser
	lAddr addr
	rAddr addr
}

// Read reads data from the connection.
func (c *conn) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

// Write writes data to the connection.
func (c *conn) Write(b []byte) (n int, err error) {
	return c.w.Write(b)
}

// Close closes the connection.
func (c *conn) Close() (err error) {
	if c.w != nil {
		err = c.w.Close()
	}
	if c.r != nil {
		err = c.r.Close()
	}
	return
}

// LocalAddr returns the local network address.
func (c *conn) LocalAddr() net.Addr {
	return c.lAddr
}

// RemoteAddr returns the remote network address.
func (c *conn) RemoteAddr() net.Addr {
	return c.rAddr
}

// SetDeadline implements the net.Conn SetDeadline method.
func (c *conn) SetDeadline(t time.Time) error {
	return errors.New("not supported")
}

// SetReadDeadline implements the net.Conn SetReadDeadline method.
func (c *conn) SetReadDeadline(t time.Time) error {
	return errors.New("not supported")
}

// SetWriteDeadline implements the net.Conn SetWriteDeadline method.
func (c *conn) SetWriteDeadline(t time.Time) error {
	return errors.New("not supported")
}

// Dial connects to an address.
func Dial(address string) (net.Conn, error) {
	rAddr := addr{network: network, address: address}
	var acceptor *acceptor
	r, w := io.Pipe()
	conn := &conn{w: w, lAddr: rAddr}
	addrs.locker.RLock()
	l, ok := addrs.listeners[rAddr]
	if !ok {
		addrs.locker.RUnlock()
		return nil, errors.New("connection refused")
	}
	addrs.locker.RUnlock()
	l.locker.Lock()
	for {
		if len(l.acceptors) > 0 {
			acceptor = l.acceptors[len(l.acceptors)-1]
			l.acceptors = l.acceptors[:len(l.acceptors)-1]
			break
		}
		l.cond.Wait()
	}
	l.locker.Unlock()
	conn.r = acceptor.reader
	conn.rAddr = conn.lAddr
	acceptor.conn.r = r
	acceptor.conn.rAddr = conn.lAddr
	close(acceptor.done)
	return conn, nil
}

// listener implements the net.Listener interface.
type listener struct {
	lAddr     addr
	cond      sync.Cond
	locker    sync.Mutex
	acceptors []*acceptor
	done      chan struct{}
	closed    int32
}

type acceptor struct {
	*conn
	reader io.ReadCloser
	done   chan struct{}
}

// Listen announces on the local address.
func Listen(address string) (net.Listener, error) {
	lAddr := addr{network: network, address: address}
	l := &listener{lAddr: lAddr, done: make(chan struct{})}
	l.cond.L = &l.locker
	addrs.locker.Lock()
	if _, ok := addrs.listeners[l.lAddr]; ok {
		addrs.locker.Unlock()
		return nil, errors.New("address already in use")
	}
	addrs.listeners[l.lAddr] = l
	addrs.locker.Unlock()
	return l, nil
}

// Accept waits for and returns the next connection to the listener.
func (l *listener) Accept() (net.Conn, error) {
	r, w := io.Pipe()
	acceptor := &acceptor{conn: &conn{w: w, lAddr: l.lAddr}, reader: r}
	acceptor.done = make(chan struct{})
	l.locker.Lock()
	l.acceptors = append(l.acceptors, acceptor)
	l.locker.Unlock()
	l.cond.Broadcast()
	select {
	case <-acceptor.done:
		return acceptor.conn, nil
	case <-l.done:
		return nil, errors.New("use of closed network connection")
	}
}

// Close closes the listener.
// Any blocked accept operations will be unblocked and return errors.
func (l *listener) Close() error {
	if atomic.CompareAndSwapInt32(&l.closed, 0, 1) {
		close(l.done)
	}
	addrs.locker.Lock()
	delete(addrs.listeners, l.lAddr)
	addrs.locker.Unlock()
	l.locker.Lock()
	acceptors := l.acceptors
	l.acceptors = nil
	l.locker.Unlock()
	l.cond.Broadcast()
	for _, acceptor := range acceptors {
		acceptor.Close()
	}
	return nil
}

// Addr returns the listener's network address.
func (l *listener) Addr() net.Addr {
	return l.lAddr
}
