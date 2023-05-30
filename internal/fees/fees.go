package fees

import (
	"bufio"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/namelew/RelativeClock/package/messages"
	"github.com/namelew/RelativeClock/package/minheap"
)

type Fees struct {
	currentTime uint64
	timeline    *minheap.MinHeap[uint64]
	lock        *sync.Mutex
	barrier     chan interface{}
}

func New() *Fees {
	return &Fees{
		currentTime: 1,
		timeline:    &minheap.MinHeap[uint64]{},
		lock:        &sync.Mutex{},
		barrier:     make(chan interface{}, 1),
	}
}

func (f *Fees) sender() {
	for {
		<-f.barrier

		f.lock.Lock()
		data, err := f.timeline.ExtractMin()

		for err == nil {
			m, ok := data.(*messages.Message)

			if !ok {
				log.Println("Corrupted data")
				data, err = f.timeline.ExtractMin()
				continue
			}

			conn, lerr := net.Dial("tcp", os.Getenv("SERVER"))

			if lerr != nil {
				f.timeline.Insert(m)
				log.Println("Unable to connect with server! ", err.Error())
				break
			}

			if lerr := m.Send(conn); lerr != nil {
				conn.Close()
				f.timeline.Insert(m)
				log.Println("Unable to send request to server! ", err.Error())
				break
			}

			time.Sleep(time.Second)

			buffer := make([]byte, 1024)

			n, lerr := conn.Read(buffer)

			if lerr != nil {
				conn.Close()
				f.timeline.Insert(data)
				log.Println("Incapaz de receber resposta do servidor! ", err.Error())
				break
			}

			var response messages.Message

			if lerr := response.Unpack(buffer[:n]); lerr != nil {
				conn.Close()
				f.timeline.Insert(data)
				log.Println("Incapaz de ler resposta do servidor! ", err.Error())
				break
			}

			conn.Close()

			if response.Action != messages.ACK {
				f.timeline.Insert(data)
				log.Println("Violação na sequência temporal!")
				break
			}

			log.Printf("Request from %d in time %d was send to server\n", m.Id, m.Timestep)

			data, err = f.timeline.ExtractMin()
		}
		f.lock.Unlock()
	}
}

func (f *Fees) handlerResquet(c net.Conn) {
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

	if in.Timestep > f.currentTime {
		log.Printf("Denying request from %d in time %d\n", in.Id, in.Timestep)
		out = messages.Message{
			Id:       0,
			Action:   messages.ERROR,
			Timestep: f.currentTime,
		}
	} else {
		f.lock.Lock()

		log.Printf("Receive request from %d in time %d\n", in.Id, in.Timestep)

		if f.currentTime == in.Timestep {
			f.barrier <- true
		}

		f.currentTime++

		f.timeline.Insert(&in)

		out = messages.Message{
			Id:       0,
			Action:   messages.ACK,
			Timestep: f.currentTime,
		}

		f.lock.Unlock()
	}

	if err := out.Send(c); err != nil {
		log.Println("Unable to send response", err.Error())
	}
}

func (f *Fees) Run() {
	l, err := net.Listen("tcp", os.Getenv("FEES"))

	if err != nil {
		log.Println("Unable to bind port. ", err.Error())
		return
	}

	go f.sender()

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Println("Unable to handler connection. ", err.Error())
			continue
		}

		go func(c net.Conn) {
			f.handlerResquet(c)
		}(conn)
	}
}
