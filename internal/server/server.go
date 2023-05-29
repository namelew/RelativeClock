package server

import (
	"bufio"
	"log"
	"net"
	"os"
	"time"

	"github.com/namelew/RelativeClock/package/messages"
	"github.com/namelew/RelativeClock/package/minheap"
)

const SYNCTIMER time.Duration = time.Second * 30

type Bank struct {
	currentTime uint64
	timeline    *minheap.MinHeap[uint64]
	value       int64
}

func New() *Bank {
	return &Bank{
		currentTime: 1,
		timeline:    minheap.New[uint64](),
		value:       1000,
	}
}

func (b *Bank) updateTime(m *messages.Message) {
	if b.currentTime >= m.Value() {
		b.currentTime++
	} else {
		b.currentTime = m.Value() - m.Id + 1
	}
}

func (b *Bank) handlerResquet(c net.Conn) {
	var in, out messages.Message
	buffer := make([]byte, 1024)

	n, err := bufio.NewReader(c).Read(buffer)

	if err != nil {
		log.Println("Unable to read data from "+c.RemoteAddr().String()+". ", err.Error())
		return
	}

	if err := in.Unpack(buffer[:n]); err != nil {
		log.Println("Unable to unpack data from "+c.RemoteAddr().String()+". ", err.Error())
		return
	}

	b.updateTime(&in)

	b.timeline.Insert(&in)

	out = messages.Message{
		Id:             0,
		Action:         messages.ACK,
		Timestep:       b.currentTime,
	}

	if err := out.Send(c); err != nil {
		log.Println("Unable to send response", err.Error())
	}
}

func (b *Bank) report() {
	d, err := b.timeline.ExtractMin()

	if err != nil {
		log.Println("History is empty")
		return
	}

	m, ok := d.(*messages.Message)

	if !ok {
		log.Println("Runtime error, corrupted message in history")
		return
	}

	conn, er := net.Dial("tcp", os.Getenv("SERVER"))

	if er != nil {
		b.timeline.Insert(d)
		log.Println("Unable to create connection with main server. ", er.Error())
		return
	}

	defer conn.Close()

	if er := m.Send(conn); er != nil {
		b.timeline.Insert(d)
		log.Println("Unable to send timeline to main server. ", er.Error())
		return
	}
}

func (b *Bank) Run() {
	l, err := net.Listen("tcp", os.Getenv("SERVER"))

	if err != nil {
		log.Println("Unable to bind port. ", err.Error())
		return
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Println("Unable to handler connection. ", err.Error())
			continue
		}

		go func(c net.Conn) {
			b.handlerResquet(c)
			b.report()
		}(conn)
	}
}
