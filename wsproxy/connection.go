package wsproxy

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

type (
	// DisconnectFunc is the callback which is fired when a client/connection closed
	DisconnectFunc func()
	// NativeMessageFunc is the callback for native websocket messages, receives one []byte parameter which is the raw client's message
	// NativeMessageFunc func([]byte)
)

type Connection struct {
	underline *websocket.Conn
	id        string
	server    *Server
	ctx       context.Context
	//started                  bool
	disconnected          bool
	onDisconnectListeners []DisconnectFunc
	//onNativeMessageListeners []NativeMessageFunc
	self     Emitter // pre-defined emitter than sends message to its self client
	writerMu sync.Mutex
}

func newConnection(ctx context.Context, s *Server, underlineConn *websocket.Conn, id string) *Connection {
	c := &Connection{
		underline: underlineConn,
		id:        id,
		ctx:       ctx,
		server:    s,
	}

	c.self = newEmitter(c, c.id)

	return c
}

// ErrAlreadyDisconnected can be reported on the `Connection#Disconnect` function whenever the caller tries to close the
// connection when it is already closed by the client or the caller previously.
var ErrAlreadyDisconnected = errors.New("already disconnected")

func (c *Connection) Disconnect() error {
	if c == nil || c.disconnected {
		return ErrAlreadyDisconnected
	}
	return c.server.Disconnect(c.ID())
}

func (c *Connection) fireDisconnect() {
	for i := range c.onDisconnectListeners {
		c.onDisconnectListeners[i]()
	}
}

func (c *Connection) OnDisconnect(cb DisconnectFunc) {
	c.onDisconnectListeners = append(c.onDisconnectListeners, cb)
}

func (c *Connection) ID() string {
	return c.id
}

func (c *Connection) Write(data []byte) error {
	// for any-case the app tries to write from different goroutines,
	// we must protect them because they're reporting that as bug...
	c.writerMu.Lock()

	// .WriteMessage same as NextWriter and close (flush)
	err := c.underline.WriteMessage(websocket.TextMessage, data)
	c.writerMu.Unlock()
	if err != nil {
		// if failed then the connection is off, fire the disconnect
		c.Disconnect()
	}
	return err
}
