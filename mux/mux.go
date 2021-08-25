package mux

import (
	"bufio"
	"crypto/tls"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/php2go/netpollmux/internal/logger"
	"github.com/php2go/netpollmux/netpoll"
)

// DefaultServer is the default HTTP server.
var DefaultServer = NewRoute()

// Route is an HTTP server.
type Route struct {
	*mux.Router
	Handler http.Handler
	// TLSConfig optionally provides a TLS configuration for use
	// by ServeTLS and ListenAndServeTLS. Note that this value is
	// cloned by ServeTLS and ListenAndServeTLS, so it's not
	// possible to modify the configuration with methods like
	// tls.Config.SetSessionTicketKeys. To use
	// SetSessionTicketKeys, use Server.Serve with a TLS Listener
	// instead.
	TLSConfig *tls.Config

	fast      bool
	poll      bool
	mut       sync.Mutex
	listeners []net.Listener
	pollers   []*netpoll.Server
}

// NewRoute returns a new NewRouter instance.
func NewRoute() *Route {
	return &Route{Router: mux.NewRouter()}
}

// SetFast enables the Server to use simple request parser.
func (m *Route) SetFast(fast bool) {
	m.fast = fast
}

// SetPoll enables the Server to use netpoll based on epoll/kqueue.
func (m *Route) SetPoll(poll bool) {
	m.poll = poll
}

// Run listens on the TCP network address addr and then calls
// Serve with m to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// Run always returns a non-nil error.
func (m *Route) Run(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	logger.Info("serve success")
	return m.Serve(ln)
}

// RunTLS is like Run but with a cert file and a key file.
func (m *Route) RunTLS(addr string, certFile, keyFile string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return m.ServeTLS(ln, certFile, keyFile)
}

// RunUnix todo
func RunUnix(file string) error {
	return nil
}

// Serve accepts incoming connections on the Listener l, creating a
// new service goroutine for each, or registering the conn fd to poll
// that will trigger the fd to read requests and then call handler
// to reply to them.
func (m *Route) Serve(l net.Listener) error {
	return m.serve(l, m.TLSConfig)
}

// ServeTLS accepts incoming connections on the Listener l, creating a
// new service goroutine for each. The service goroutines perform TLS
// setup and then read requests, calling srv.Handler to reply to them.
//
// Files containing a certificate and matching private key for the
// server must be provided if neither the Server's
// TLSConfig.Certificates nor TLSConfig.GetCertificate are populated.
// If the certificate is signed by a certificate authority, the
// certFile should be the concatenation of the server's certificate,
// any intermediates, and the CA's certificate.
//
// ServeTLS always returns a non-nil error. After Shutdown or Close, the
// returned error is ErrServerClosed.
func (m *Route) ServeTLS(l net.Listener, certFile, keyFile string) error {
	config := m.TLSConfig
	if config == nil {
		config = &tls.Config{}
	}
	if !strSliceContains(config.NextProtos, "http/1.1") {
		config.NextProtos = append(config.NextProtos, "http/1.1")
	}
	configHasCert := len(config.Certificates) > 0 || config.GetCertificate != nil
	if !configHasCert || certFile != "" || keyFile != "" {
		var err error
		config.Certificates = make([]tls.Certificate, 1)
		config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return err
		}
	}
	return m.serve(l, config)
}

func (m *Route) serve(l net.Listener, config *tls.Config) error {
	if m.poll {
		var handler = m.Handler
		if handler == nil {
			handler = m
		}
		var h = netpoll.NewConHandler()
		type Context struct {
			reader  *bufio.Reader
			rw      *bufio.ReadWriter
			conn    net.Conn
			serving sync.Mutex
		}
		h.SetUpgrade(func(conn net.Conn) (netpoll.Context, error) {
			if config != nil {
				tlsConn := tls.Server(conn, config)
				if err := tlsConn.Handshake(); err != nil {
					conn.Close()
					return nil, err
				}
				conn = tlsConn
			}
			reader := bufio.NewReader(conn)
			rw := bufio.NewReadWriter(reader, bufio.NewWriter(conn))
			return &Context{reader: reader, conn: conn, rw: rw}, nil
		})
		if m.fast {
			h.SetServe(func(context netpoll.Context) error {
				ctx := context.(*Context)
				var err error
				var req *http.Request
				ctx.serving.Lock()
				req, err = ReadFastRequest(ctx.reader)
				if err != nil {
					ctx.serving.Unlock()
					return err
				}
				res := NewResponse(req, ctx.conn, ctx.rw)
				handler.ServeHTTP(res, req)
				res.FinishRequest()
				ctx.serving.Unlock()
				FreeRequest(req)
				FreeResponse(res)
				return nil
			})
		} else {
			h.SetServe(func(context netpoll.Context) error {
				ctx := context.(*Context)
				var err error
				var req *http.Request
				ctx.serving.Lock()
				req, err = http.ReadRequest(ctx.reader)
				if err != nil {
					ctx.serving.Unlock()
					return err
				}
				res := NewResponse(req, ctx.conn, ctx.rw)
				handler.ServeHTTP(res, req)
				res.FinishRequest()
				ctx.serving.Unlock()
				FreeResponse(res)
				return nil
			})
		}
		poller := &netpoll.Server{
			Handler: h,
		}
		m.mut.Lock()
		m.pollers = append(m.pollers, poller)
		m.mut.Unlock()
		return poller.Serve(l)
	}
	if config != nil {
		l = tls.NewListener(l, config)
	}
	m.mut.Lock()
	m.listeners = append(m.listeners, l)
	m.mut.Unlock()
	if m.fast {
		for {
			conn, err := l.Accept()
			if err != nil {
				return err
			}
			go m.serveFastConn(conn)
		}
	} else {
		for {
			conn, err := l.Accept()
			if err != nil {
				return err
			}
			go m.serveConn(conn)
		}
	}
}

// Close closes the HTTP server.
func (m *Route) Close() error {
	m.mut.Lock()
	defer m.mut.Unlock()
	for _, lis := range m.listeners {
		lis.Close()
	}
	m.listeners = []net.Listener{}
	for _, poller := range m.pollers {
		poller.Close()
	}
	m.pollers = []*netpoll.Server{}
	m.Handler = nil
	return nil
}

func (m *Route) serveConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	rw := bufio.NewReadWriter(reader, bufio.NewWriter(conn))
	var err error
	var req *http.Request
	var handler = m.Handler
	if handler == nil {
		handler = m
	}
	for {
		req, err = http.ReadRequest(reader)
		if err != nil {
			break
		}
		res := NewResponse(req, conn, rw)
		handler.ServeHTTP(res, req)
		res.FinishRequest()
		FreeResponse(res)
	}
}

func (m *Route) serveFastConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	rw := bufio.NewReadWriter(reader, bufio.NewWriter(conn))
	var err error
	var req *http.Request
	var handler = m.Handler
	if handler == nil {
		handler = m
	}
	for {
		req, err = ReadFastRequest(reader)
		if err != nil {
			break
		}
		res := NewResponse(req, conn, rw)
		handler.ServeHTTP(res, req)
		res.FinishRequest()
		FreeRequest(req)
		FreeResponse(res)
	}
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve with handler to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// The handler is typically nil, in which case the DefaultServeMux is used.
//
// ListenAndServe always returns a non-nil error.
func ListenAndServe(addr string, handler http.Handler) error {
	rum := DefaultServer
	rum.Handler = handler
	return rum.Run(addr)
}

// ListenAndServeTLS acts identically to ListenAndServe, except that it
// expects HTTPS connections. Additionally, files containing a certificate and
// matching private key for the server must be provided. If the certificate
// is signed by a certificate authority, the certFile should be the concatenation
// of the server's certificate, any intermediates, and the CA's certificate.
func ListenAndServeTLS(addr, certFile, keyFile string, handler http.Handler) error {
	rum := DefaultServer
	rum.Handler = handler
	return rum.RunTLS(addr, certFile, keyFile)
}

func strSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}
