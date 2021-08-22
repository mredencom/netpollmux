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
	laddr addr
	raddr addr
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
	return c.laddr
}

// RemoteAddr returns the remote network address.
func (c *conn) RemoteAddr() net.Addr {
	return c.raddr
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
	raddr := addr{network: network, address: address}
	var accepter *accepter
	r, w := io.Pipe()
	conn := &conn{w: w, laddr: raddr}
	addrs.locker.RLock()
	l, ok := addrs.listeners[raddr]
	if !ok {
		addrs.locker.RUnlock()
		return nil, errors.New("connection refused")
	}
	addrs.locker.RUnlock()
	l.locker.Lock()
	for {
		if len(l.accepters) > 0 {
			accepter = l.accepters[len(l.accepters)-1]
			l.accepters = l.accepters[:len(l.accepters)-1]
			break
		}
		l.cond.Wait()
	}
	l.locker.Unlock()
	conn.r = accepter.reader
	conn.raddr = conn.laddr
	accepter.conn.r = r
	accepter.conn.raddr = conn.laddr
	close(accepter.done)
	return conn, nil
}

// listener implements the net.Listener interface.
type listener struct {
	laddr     addr
	cond      sync.Cond
	locker    sync.Mutex
	accepters []*accepter
	done      chan struct{}
	closed    int32
}

type accepter struct {
	*conn
	reader io.ReadCloser
	done   chan struct{}
}

// Listen announces on the local address.
func Listen(address string) (net.Listener, error) {
	laddr := addr{network: network, address: address}
	l := &listener{laddr: laddr, done: make(chan struct{})}
	l.cond.L = &l.locker
	addrs.locker.Lock()
	if _, ok := addrs.listeners[l.laddr]; ok {
		addrs.locker.Unlock()
		return nil, errors.New("address already in use")
	}
	addrs.listeners[l.laddr] = l
	addrs.locker.Unlock()
	return l, nil
}

// Accept waits for and returns the next connection to the listener.
func (l *listener) Accept() (net.Conn, error) {
	r, w := io.Pipe()
	accepter := &accepter{conn: &conn{w: w, laddr: l.laddr}, reader: r}
	accepter.done = make(chan struct{})
	l.locker.Lock()
	l.accepters = append(l.accepters, accepter)
	l.locker.Unlock()
	l.cond.Broadcast()
	select {
	case <-accepter.done:
		return accepter.conn, nil
	case <-l.done:
		return nil, errors.New("use of closed network connection")
	}
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (l *listener) Close() error {
	if atomic.CompareAndSwapInt32(&l.closed, 0, 1) {
		close(l.done)
	}
	addrs.locker.Lock()
	delete(addrs.listeners, l.laddr)
	addrs.locker.Unlock()
	l.locker.Lock()
	accepters := l.accepters
	l.accepters = nil
	l.locker.Unlock()
	l.cond.Broadcast()
	for _, accepter := range accepters {
		accepter.Close()
	}
	return nil
}

// Addr returns the listener's network address.
func (l *listener) Addr() net.Addr {
	return l.laddr
}
