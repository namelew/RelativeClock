package database

import (
	"bufio"
	"log"
	"net"
	"os"
	"sync"

	"github.com/namelew/RelativeClock/package/messages"
	"github.com/namelew/RelativeClock/package/minheap"
)

type Database struct {
	currentTime uint64
	value       float64
	pipeline    *minheap.MinHeap[uint64]
	lock        *sync.Mutex
	barrier     chan interface{}
}

const STARTVALUE float64 = 1000

func New() *Database {
	return &Database{
		currentTime: 1,
		value:       STARTVALUE,
		lock:        &sync.Mutex{},
		pipeline:    &minheap.MinHeap[uint64]{},
		barrier:     make(chan interface{}, 1),
	}
}

func (d *Database) worker() {
	for {
		<-d.barrier
		d.lock.Lock()
		data, err := d.pipeline.ExtractMin()

		for err != nil {
			m, ok := data.(*messages.Message)

			if !ok {
				log.Println("Corrupted data")
				data, err = d.pipeline.ExtractMin()
				continue
			}

			switch m.Action {
			case messages.DEP:
				log.Printf("Running request from Warehouse, increament %f in timestep %d\n", m.Payload, m.Timestep)
				d.value += m.Payload
			case messages.FEE:
				log.Printf("Running request from Fees, apply fee of %f in timestep %d\n", m.Payload, m.Timestep)
				d.value *= m.Payload
			}

			data, err = d.pipeline.ExtractMin()
		}
		d.lock.Unlock()
	}
}

func (d *Database) Run() {
	l, err := net.Listen("tcp", os.Getenv("SERVER"))

	if err != nil {
		log.Println("Unable to bind port. ", err.Error())
		return
	}

	go d.worker()

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

			if err != nil {
				log.Println("Unable to read data from "+c.RemoteAddr().String()+". ", err.Error())
				return
			}

			if err := in.Unpack(buffer[:n]); err != nil {
				log.Println("Unable to unpack data from "+c.RemoteAddr().String()+". ", err.Error())
				return
			}

			if in.Timestep > d.currentTime {
				out = messages.Message{
					Id:       0,
					Action:   messages.ERROR,
					Timestep: d.currentTime,
				}
			} else {
				d.lock.Lock()

				d.pipeline.Insert(&in)

				if d.currentTime == in.Timestep {
					d.barrier <- true
				}

				d.currentTime++

				out = messages.Message{
					Id:       0,
					Action:   messages.ACK,
					Timestep: d.currentTime,
				}

				d.lock.Unlock()
			}

			if err := out.Send(c); err != nil {
				log.Println("Unable to send response.", err.Error())
			}
		}(conn)
	}
}
