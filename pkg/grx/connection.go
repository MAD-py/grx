package grx

import (
	"bytes"
	"encoding/gob"
	"net"
)

type Action uint8

const (
	stop Action = iota
	run
	kill
)

type Message struct {
	Action Action
	Params interface{}
}

type Response struct {
	Message string
	Body    map[string]interface{}
}

type connection struct {
	conn net.Conn
}

func (c *connection) read() (*Message, bool) {
	buf := make([]byte, 1024)
	_, err := c.conn.Read(buf)
	if err != nil {
		c.write(err.Error(), nil)
		return nil, false
	}

	message := &Message{}
	gob.NewDecoder(bytes.NewBuffer(buf)).Decode(message)
	return message, true
}

func (c *connection) write(message string, body map[string]interface{}) {
	res := &Response{
		Message: message,
		Body:    body,
	}

	buf := &bytes.Buffer{}
	gob.NewEncoder(buf).Encode(res)
	c.conn.Write(buf.Bytes())
}

func (c *connection) close() {
	c.conn.Close()
}

func newConn(conn net.Conn) *connection {
	return &connection{
		conn: conn,
	}
}
