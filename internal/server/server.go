package server

import (
	"bufio"
	"log"
	"net"
	"os"

	"github.com/namelew/RelativeClock/package/messages"
)

type Bank struct {
	currentTime uint64
	auxServer   bool
}

func New() *Bank {
	return &Bank{
		currentTime: 0,
		auxServer:   false,
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

			switch in.Action {
			case messages.ADD:
			case messages.SUB:
			case messages.REP:
				if b.auxServer {
					return
				}
			}

			if err := out.Send(c); err != nil {
				log.Println(err.Error())
			}
		}(conn)
	}
}
