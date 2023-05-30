package messages

import (
	"encoding/json"
	"errors"
	"net"
)

type Action uint32

const (
	ERROR Action = 0
	DEP   Action = 1
	FEE   Action = 2
	ACK   Action = 3
)

type Message struct {
	Id       uint64
	Action   Action
	Timestep uint64
	Payload  float64
}

func (m *Message) Pack() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unpack(b []byte) error {
	return json.Unmarshal(b, m)
}

func (m *Message) Value() uint64 {
	return m.Timestep + m.Id
}

func (m *Message) Send(c net.Conn) error {
	payload, err := m.Pack()

	if err != nil {
		return errors.New("Unable to pack message. " + err.Error())
	}

	_, err = c.Write(payload)

	if err != nil {
		return errors.New("Unable to send message to " + c.RemoteAddr().String() + ". " + err.Error())
	}

	return err
}
