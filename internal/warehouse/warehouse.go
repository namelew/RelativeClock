package server

import (
	"bufio"
	"log"
	"net"
	"os"
	"sync"

	"github.com/namelew/RelativeClock/package/messages"
	"github.com/namelew/RelativeClock/package/minheap"
)

type Warehouse struct {
	currentTime uint64
	timeline    *minheap.MinHeap[uint64]
	lock        *sync.Mutex
}

func New() *Warehouse {
	return &Warehouse{
		currentTime: 1,
		timeline:    &minheap.MinHeap[uint64]{},
		lock:        &sync.Mutex{},
	}
}

func (w *Warehouse) handlerResquet(c net.Conn) {
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

	if in.Timestep > w.currentTime {
		out = messages.Message{
			Id:       0,
			Action:   messages.ERROR,
			Timestep: w.currentTime,
		}
	} else {
		w.lock.Lock()

		w.currentTime++

		w.timeline.Insert(&in)

		out = messages.Message{
			Id:       0,
			Action:   messages.ACK,
			Timestep: w.currentTime,
		}

		w.lock.Unlock()
	}

	if err := out.Send(c); err != nil {
		log.Println("Unable to send response", err.Error())
	}
}

func (w *Warehouse) Run() {
	l, err := net.Listen("tcp", os.Getenv("WAREROUSE"))

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
			w.handlerResquet(c)
		}(conn)
	}
}
