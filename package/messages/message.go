package messages

import "encoding/json"

type Action uint32

const (
	ERROR Action = 0
	ADD   Action = 1
	SUB   Action = 2
	REP   Action = 3
)

type Message struct {
	Id       uint64
	Action   Action
	Timestep uint64
}

func (m *Message) Pack() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Message) Unpack(b []byte) error {
	return json.Unmarshal(b, m)
}
