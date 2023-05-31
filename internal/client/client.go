package client

import (
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/namelew/RelativeClock/package/messages"
	"github.com/namelew/RelativeClock/package/minheap"
)

type Client struct {
	id           uint64
	currentTime  uint64
	pipeline     *minheap.MinHeap[uint64]
	neighborhood []string
	lock         sync.Mutex
}

const script string = "./script.in"

func removeBackslashChars(s string) string {
	var result strings.Builder
	skip := false
	for _, r := range s {
		if skip {
			skip = false
			continue
		}
		if r == '\\' {
			skip = true
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

func (c *Client) getNeighbors() {
	c.neighborhood = strings.Split(os.Getenv("NEIGHBORS"), ",")
}

func (c *Client) readScript() {
	data, err := os.ReadFile(script)

	if err != nil {
		log.Fatal("Unable to open script file!", err.Error())
	}

	lines := strings.Split(string(data), "\n")
	re := regexp.MustCompile(`(^\d+) (\w) (\d+)$`)

	for i := range lines {
		line := removeBackslashChars(lines[i])

		match := re.FindStringSubmatch(line)

		if match != nil {
			tms, err := strconv.ParseInt(match[1], 10, 64)

			if err != nil {
				log.Fatal("Unable to parser message timestep!", err.Error())
			}

			act := messages.ERROR

			switch match[2] {
			case "D":
				act = messages.DEP
			case "J":
				act = messages.FEE
			}

			value, err := strconv.ParseFloat(match[3], 64)

			if err != nil {
				log.Fatal("Unable to parser message payload!", err.Error())
			}

			c.pipeline.Insert(&messages.Message{
				Id:       c.id,
				Action:   act,
				Timestep: uint64(tms),
				Payload:  value,
			})
		}
	}
}

func (c *Client) handler() {
	l, err := net.Listen("tcp", c.neighborhood[c.id-1])

	if err != nil {
		log.Fatal("Unable to bind port!", err.Error())
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Println("Unable to accept connection!", err.Error())
			continue
		}

		go func(connection net.Conn) {
			var request messages.Message

			if err := request.Receive(connection); err != nil {
				log.Println("Unable to receive message!", err.Error())
				return
			}

			c.lock.Lock()
			defer c.lock.Unlock()

			if request.Value() <= c.currentTime {
				response := messages.Message{
					Id:       c.id,
					Action:   messages.ACK,
					Timestep: c.currentTime,
				}

				if err := response.Send(connection); err != nil {
					log.Println("Unable to send response!", err.Error())
					return
				}
			}
		}(conn)
	}
}

func New(id uint64) *Client {
	return &Client{
		id:          id,
		currentTime: 1,
		pipeline:    &minheap.MinHeap[uint64]{},
	}
}

func (c *Client) Run() {
	c.getNeighbors()
	c.readScript()

	c.handler()
}
