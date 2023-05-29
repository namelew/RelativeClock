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
	auxServer   bool
	timeline    *minheap.MinHeap[uint64]
	value       int64
}

func New() *Bank {
	return &Bank{
		currentTime: 1,
		auxServer:   false,
		timeline:    minheap.New[uint64](),
		value:       0,
	}
}

func (b *Bank) updateTime(m *messages.Message) {
	if b.currentTime >= m.Value() {
		b.currentTime++
	} else {
		b.currentTime = m.Value() - m.Id + 1
	}
}

func (b *Bank) sync() {
	for {
		<-time.After(SYNCTIMER)

		d, err := b.timeline.ExtractMin()

		for err == nil {
			m, ok := d.(*messages.Message)

			if !ok {
				log.Println("Unable to read message")
				continue
			}

			switch m.Action {
			case messages.DEP:
				log.Printf("Running request from %d into time %d, adding %d into balance\n", m.Id, m.Value()-m.Id, m.Payload)
				b.value += m.Payload
			case messages.FEE:
				log.Printf("Running request from %d into time %d, decressing %d into the balance\n", m.Id, m.Value()-m.Id, m.Payload)
				b.value -= m.Payload
			}

			d, err = b.timeline.ExtractMin()
		}

		log.Println("Finishing syncronization. Current balance: ", b.value)
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

	switch in.Action {
	case messages.STM:
		if b.auxServer {
			b.currentTime = in.Timestep
		}
	default:
		b.timeline.Insert(&in)
		b.updateTime(&in)

		out = messages.Message{
			Id:             0,
			Action:         messages.ACK,
			Timestep:       b.currentTime,
			SenderTimestep: b.currentTime,
		}

		if err := out.Send(c); err != nil {
			log.Println("Unable to send response", err.Error())
		}

		go func() {
			conn, err := net.Dial("tcp", os.Getenv("SERVERAUX"))

			if err != nil {
				log.Println("Unable to create connection with proxy! ", err.Error())
				return
			}

			defer conn.Close()

			m := messages.Message{
				Action: messages.STM,
				Timestep: b.currentTime,
				SenderTimestep: b.currentTime,
			}

			if err := m.Send(conn); err != nil {
				log.Println("Unable to send timestep to proxy! ", err.Error())
				return
			}
		}()
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
		log.Println("Starting aux server")
		l, err = net.Listen("tcp", os.Getenv("SERVERAUX"))

		if err != nil {
			log.Println("Unable to bind port. ", err.Error())
			return
		}
		b.auxServer = true
	}

	if !b.auxServer {
		go b.sync()
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Println("Unable to handler connection. ", err.Error())
			continue
		}

		go func(c net.Conn) {
			b.handlerResquet(c)
			if b.auxServer {
				b.report()
			}
		}(conn)
	}
}
