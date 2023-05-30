package fees

import (
	"bufio"
	"log"
	"net"
	"os"
	"sync"

	"github.com/namelew/RelativeClock/package/messages"
	"github.com/namelew/RelativeClock/package/minheap"
)

type Fees struct {
	currentTime uint64
	timeline    *minheap.MinHeap[uint64]
	lock        *sync.Mutex
}

func New() *Fees {
	return &Fees{
		currentTime: 1,
		timeline:    &minheap.MinHeap[uint64]{},
		lock:        &sync.Mutex{},
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
