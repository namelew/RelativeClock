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

const SYNCTIMER time.Duration = time.Second * 3

type Bank struct {
	currentTime uint64
	auxServer   bool
	timeline    *minheap.MinHeap[uint64]
	value       int64
}

func New() *Bank {
	return &Bank{
		currentTime: 0,
		auxServer:   false,
		timeline:    minheap.New[uint64](),
		value:       0,
	}
}

func (b *Bank) sync() {
	if !b.auxServer {
		for {
			<-time.After(SYNCTIMER)

			d, err := b.timeline.ExtractMin()

			for err != nil {
				m, ok := d.(*messages.Message)

				if !ok {
					log.Println("Unable to read message")
					continue
				}

				switch m.Action {
				case messages.ADD:
					log.Printf("Running request from %d, adding %d into balance\n", m.Id, m.Payload)
					b.value += m.Payload
				case messages.SUB:
					log.Printf("Running request from %d, decressing %d into balance\n", m.Id, m.Payload)
					b.value -= m.Payload
				}

				d, err = b.timeline.ExtractMin()
			}

			log.Println("Finishing syncronization. Current value: ", b.value)
		}
	} else {
		for {
			<-time.After(SYNCTIMER - (time.Second / 2))

			d, err := b.timeline.ExtractMin()

			for err != nil {
				m, ok := d.(*messages.Message)

				if !ok {
					log.Println("Unable to read message")
					continue
				}

				conn, er := net.Dial("tcp", os.Getenv("SERVER"))

				if er != nil {
					b.timeline.Insert(m)
					log.Println("Unable to read create connection with main server. ", er.Error())
					continue
				}

				if er := m.Send(conn); er != nil {
					b.timeline.Insert(m)
					log.Println("Unable to send timeline to main server. ", er.Error())
					conn.Close()
					continue
				}

				conn.Close()

				d, err = b.timeline.ExtractMin()
			}
		}
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

	go b.sync()

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Println("Unable to handler connection. ", err.Error())
			continue
		}

		go func(c net.Conn) {
			var in, out messages.Message
			buffer := make([]byte, 1024)

			n, err := bufio.NewReader(c).Read(buffer)

			defer c.Close()

			if err != nil {
				log.Println("Unable to read data from "+c.RemoteAddr().String()+". ", err.Error())
				return
			}

			if err := in.Unpack(buffer[:n]); err != nil {
				log.Println("Unable to unpack data from "+c.RemoteAddr().String()+". ", err.Error())
				return
			}

			b.timeline.Insert(&in)

			out = messages.Message{
				Id:             0,
				Action:         messages.ACK,
				Timestep:       b.currentTime,
				SenderTimestep: b.currentTime,
			}

			if err := out.Send(c); err != nil {
				log.Println(err.Error())
			}
		}(conn)
	}
}
