package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/namelew/RelativeClock/package/messages"
)

type Teller struct {
	id          uint64
	currentTime uint64
	server      string
}

func New(id uint64) *Teller {
	return &Teller{
		id:          id,
		currentTime: 1,
		server:      os.Getenv("SERVER"),
	}
}

func (t *Teller) Run() {
	r := bufio.NewReader(os.Stdin)

	fmt.Printf("Action\n1 - Deposito\n2 - Juros\n")
	fmt.Println("Expect: Action Value")

	for {
		fmt.Print("\nOperation: ")
		p, err := r.ReadSlice('\n')

		if err != nil {
			log.Println(err.Error())
			continue
		}

		temp := strings.ReplaceAll(string(p), "\r", "")
		temp = strings.ReplaceAll(temp, "\n", "")

		input := strings.Split(temp, " ")

		if len(input) < 2 {
			log.Println("Menos argumentos do que necessário")
			continue
		}

		action, err := strconv.Atoi(input[0])

		if err != nil {
			log.Println("Formato inválido! ", err.Error())
			continue
		}

		payload, err := strconv.ParseFloat(input[1], 64)

		if err != nil {
			log.Println("Formato inválido! ", err.Error())
			continue
		}

		m := messages.Message{
			Id:       t.id,
			Action:   messages.Action(action),
			Timestep: t.currentTime,
			Payload:  payload,
		}

		var server string
		switch m.Action {
		case messages.DEP:
			server = os.Getenv("WAREROUSE")
		case messages.FEE:
			server = os.Getenv("FEES")
		}

		conn, err := net.Dial("tcp", server)

		if err != nil {
			log.Println("Incapaz de estabelecer conexão com o servidor! ", err.Error())
			continue
		}

		if err := m.Send(conn); err != nil {
			conn.Close()
			log.Println("Incapaz de enviar requisição ao servidor! ", err.Error())
			continue
		}

		time.Sleep(time.Second)

		buffer := make([]byte, 1024)

		n, err := conn.Read(buffer)

		if err != nil {
			conn.Close()
			log.Println("Incapaz de receber resposta do servidor! ", err.Error())
			continue
		}

		if err := m.Unpack(buffer[:n]); err != nil {
			conn.Close()
			log.Println("Incapaz de ler resposta do servidor! ", err.Error())
			return
		}

		conn.Close()

		if m.Action != messages.ACK {
			log.Println("Violação na sequência temporal!")
			continue
		}

		t.currentTime++
	}
}
