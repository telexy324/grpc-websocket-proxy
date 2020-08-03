package wsproxy

type Emitter interface {
	// EmitMessage sends a native websocket message
	EmitMessage([]byte) error
}

type emitter struct {
	conn *Connection
	to   string
}

func newEmitter(c *Connection, to string) *emitter {
	return &emitter{conn: c, to: to}
}

func (e *emitter) EmitMessage(nativeMessage []byte) error {
	e.conn.server.emitMessage(e.conn.id, e.to, nativeMessage)
	return nil
}
