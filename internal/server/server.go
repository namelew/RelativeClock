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
		b.currentTime += m.Value() - m.Id + 1
	}
}

func (b *Bank) sync() {
	if !b.auxServer {
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
	} else {
		for {
			<-time.After(SYNCTIMER - (time.Second / 2))

			d, err := b.timeline.ExtractMin()

			for err == nil {
				buffer := make([]byte, 1024)
				m, ok := d.(*messages.Message)

				if !ok {
					log.Println("Unable to read message")
					continue
				}

				conn, er := net.Dial("tcp", os.Getenv("SERVER"))

				if er != nil {
					b.timeline.Insert(d)
					log.Println("Unable to read create connection with main server. ", er.Error())
					continue
				}

				if er := m.Send(conn); er != nil {
					b.timeline.Insert(d)
					log.Println("Unable to send timeline to main server. ", er.Error())
					conn.Close()
					continue
				}

				time.Sleep(time.Second / 4)

				n, er := conn.Read(buffer)

				conn.Close()

				if er != nil {
					b.timeline.Insert(d)
					log.Println("Unable to read response from main server. ", er.Error())
					continue
				}

				response := messages.Message{}

				if er := response.Unpack(buffer[:n]); er != nil {
					b.timeline.Insert(d)
					log.Println("Unable to read response from main server. ", er.Error())
					continue
				}

				b.currentTime = response.SenderTimestep

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
		if b.auxServer {
			go func() {
				for {
					time.Sleep(time.Second)
					log.Println("System current time", b.currentTime)
				}
			}()
		}
		conn, err := l.Accept()

		if err != nil {
			log.Println("Unable to handler connection. ", err.Error())
			continue
		}

		go func(c net.Conn) {
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

			b.timeline.Insert(&in)

			b.updateTime(&in)

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
