package wsproxy

import (
	"context"
	"github.com/gorilla/websocket"
	"sync"
)

type ConnectionFunc func(*Connection)

type Server struct {
	IDGenerator           func() string
	connections           sync.Map // key = the Connection ID.
	onConnectionListeners []ConnectionFunc
}

func New() *Server {
	return &Server{
		IDGenerator:           DefaultIDGenerator,
		connections:           sync.Map{}, // ready-to-use, this is not necessary.
		onConnectionListeners: make([]ConnectionFunc, 0),
	}
}

func (s *Server) Handle(ctx context.Context, websocketConn *websocket.Conn) error {
	c, err := s.HandleConnection(ctx, websocketConn)
	if err != nil {
		return err
	}
	for i := range s.onConnectionListeners {
		s.onConnectionListeners[i](c)
	}
	return nil
}

func (s *Server) addConnection(c *Connection) {
	s.connections.Store(c.id, c)
}

func (s *Server) getConnection(connID string) (*Connection, bool) {
	if cValue, ok := s.connections.Load(connID); ok {
		// this cast is not necessary,
		// we know that we always save a connection, but for good or worse let it be here.
		if conn, ok := cValue.(*Connection); ok {
			return conn, ok
		}
	}

	return nil, false
}

// wrapConnection wraps an underline connection to an iris websocket connection.
// It does NOT starts its writer, reader and event mux, the caller is responsible for that.
func (s *Server) HandleConnection(ctx context.Context, websocketConn *websocket.Conn) (*Connection, error) {
	// use the config's id generator (or the default) to create a websocket client/connection id
	cid := s.IDGenerator()
	// create the new connection
	c := newConnection(ctx, s, websocketConn, cid)
	// add the connection to the Server's list
	s.addConnection(c)

	return c, nil
}

func (s *Server) OnConnection(cb ConnectionFunc) {
	s.onConnectionListeners = append(s.onConnectionListeners, cb)
}

func (s *Server) GetConnection(connID string) *Connection {
	conn, ok := s.getConnection(connID)
	if !ok {
		return nil
	}

	return conn
}

func (s *Server) Disconnect(connID string) (err error) {
	// remove the connection from the list.
	if conn, ok := s.getConnection(connID); ok {
		conn.disconnected = true
		// fire the disconnect callbacks, if any.
		conn.fireDisconnect()
		// close the underline connection and return its error, if any.
		err = conn.underline.Close()

		s.connections.Delete(connID)
	}

	return
}

