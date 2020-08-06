package wsproxy

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"net"
	"sync"
	"time"
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

func (c *Connection) Write(websocketMessageType int, data []byte) error {
	// for any-case the app tries to write from different goroutines,
	// we must protect them because they're reporting that as bug...
	c.writerMu.Lock()

	// .WriteMessage same as NextWriter and close (flush)
	err := c.underline.WriteMessage(websocketMessageType, data)
	c.writerMu.Unlock()
	if err != nil {
		// if failed then the connection is off, fire the disconnect
		c.Disconnect()
	}
	return err
}

const (
	// WriteWait is 1 second at the internal implementation,
	// same as here but this can be changed at the future*
	WriteWait  = 1 * time.Second
	PingPeriod = 60 * 10 * time.Second / 9
)

func (c *Connection) startPinger() {

	// this is the default internal handler, we just change the writeWait because of the actions we must do before
	// the server sends the ping-pong.

	pingHandler := func(message string) error {
		err := c.underline.WriteControl(websocket.PongMessage, []byte(message), time.Now().Add(WriteWait))
		if err == websocket.ErrCloseSent {
			return nil
		} else if e, ok := err.(net.Error); ok && e.Temporary() {
			return nil
		}
		return err
	}

	c.underline.SetPingHandler(pingHandler)

	go func() {
		for {
			// using sleep avoids the ticker error that causes a memory leak
			time.Sleep(PingPeriod)
			if c.disconnected {
				// verifies if already disconected
				break
			}
			// try to ping the client, if failed then it disconnects
			err := c.Write(websocket.PingMessage, []byte{})
			if err != nil {
				// must stop to exit the loop and finish the go routine
				break
			}
		}
	}()
}
