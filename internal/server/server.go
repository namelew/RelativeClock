package server

import (
	"log"
	"net"
	"os"
)

type Bank struct {
	currentTime uint64
}

func New() *Bank {
	return &Bank{
		currentTime: 0,
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
	}

	for {
		l.Accept()
		break
	}
}
