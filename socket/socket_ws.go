package socket

import (
	"crypto/tls"
	"net"
	"netpollmux"
	"netpollmux/websocket"
)

const (
	//WSPath defines the ws path.
	WSPath = "/"
)

// WS implements the Socket interface.
type WS struct {
	Config *tls.Config
}

// WSConn implements the Conn interface.
type WSConn struct {
	*websocket.Conn
}

// Messages returns a new Messages.
func (c *WSConn) Messages() Messages {
	return c.Conn
}

// Connection returns the net.Conn.
func (c *WSConn) Connection() net.Conn {
	return c.Conn
}

// NewWSSocket returns a new WS socket.
func NewWSSocket(config *tls.Config) Socket {
	return &WS{Config: config}
}

// Scheme returns the socket's scheme.
func (t *WS) Scheme() string {
	if t.Config == nil {
		return "ws"
	}
	return "wss"
}

// Dial connects to an address.
func (t *WS) Dial(address string) (Conn, error) {
	var err error
	conn, err := websocket.Dial("tcp", address, WSPath, t.Config)
	if err != nil {
		return nil, err
	}
	return &WSConn{conn}, err
}

// Listen announces on the local address.
func (t *WS) Listen(address string) (Listener, error) {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}
	return &WSListener{l: lis, config: t.Config}, nil
}

// WSListener implements the Listener interface.
type WSListener struct {
	l      net.Listener
	server *netpollmux.Server
	config *tls.Config
}

// Accept waits for and returns the next connection to the listener.
func (l *WSListener) Accept() (Conn, error) {
	conn, err := l.l.Accept()
	if err != nil {
		return nil, err
	}
	if l.config == nil {

	}
	ws, err := websocket.Upgrade(conn, l.config)
	if err != nil {
		return nil, err
	}
	return &WSConn{ws}, err
}

// Serve serves the netpoll.Handler by the netpoll.
func (l *WSListener) Serve(handler netpollmux.Handler) error {
	if handler == nil {
		return ErrHandler
	}
	l.server = &netpollmux.Server{
		Handler: handler,
	}
	return l.server.Serve(l.l)
}

// ServeData serves the opened func and the serve func by the netpoll.
func (l *WSListener) ServeData(opened func(net.Conn) error, serve func(req []byte) (res []byte)) error {
	if opened == nil {
		return ErrOpened
	} else if serve == nil {
		return ErrServe
	}
	Upgrade := func(conn net.Conn) (netpollmux.Context, error) {
		messages, err := websocket.Upgrade(conn, l.config)
		if err != nil {
			conn.Close()
			return nil, err
		}
		opened(messages)
		return messages, nil
	}
	Serve := func(context netpollmux.Context) error {
		ws := context.(*websocket.Conn)
		msg, err := ws.ReadMessage(nil)
		if err != nil {
			return err
		}
		res := serve(msg)
		if len(res) == 0 {
			return nil
		}
		return ws.WriteMessage(res)
	}
	l.server = &netpollmux.Server{
		Handler: netpollmux.NewHandler(Upgrade, Serve),
	}
	return l.server.Serve(l.l)
}

// ServeConn serves the opened func and the serve func by the netpoll.
func (l *WSListener) ServeConn(opened func(net.Conn) (Context, error), serve func(Context) error) error {
	if opened == nil {
		return ErrOpened
	} else if serve == nil {
		return ErrServe
	}
	Upgrade := func(conn net.Conn) (netpollmux.Context, error) {
		messages, err := websocket.Upgrade(conn, l.config)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return opened(messages)
	}
	Serve := func(context netpollmux.Context) error {
		return serve(context)
	}
	l.server = &netpollmux.Server{
		Handler: netpollmux.NewHandler(Upgrade, Serve),
	}
	return l.server.Serve(l.l)
}

// ServeMessages serves the opened func and the serve func by the netpoll.\
func (l *WSListener) ServeMessages(opened func(Messages) (Context, error), serve func(Context) error) error {
	if opened == nil {
		return ErrOpened
	} else if serve == nil {
		return ErrServe
	}
	Upgrade := func(conn net.Conn) (netpollmux.Context, error) {
		messages, err := websocket.Upgrade(conn, l.config)
		if err != nil {
			conn.Close()
			return nil, err
		}
		return opened(messages)
	}
	Serve := func(context netpollmux.Context) error {
		return serve(context)
	}
	l.server = &netpollmux.Server{
		Handler: netpollmux.NewHandler(Upgrade, Serve),
	}
	return l.server.Serve(l.l)
}

// Close closes the listener.
func (l *WSListener) Close() error {
	if l.server != nil {
		return l.server.Close()
	}
	return l.l.Close()
}

// Addr returns the listener's network address.
func (l *WSListener) Addr() net.Addr {
	return l.l.Addr()
}
