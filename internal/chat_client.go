package internal

import (
	"fmt"
	"io"
	"log"
)

type ClientEventHandler struct {
	OnRead   func(io.Reader)
	OnWrite  func(io.Writer)
	OnHangUp func()
	OnErr    func()
}

func NewNoOpClientEventHandler() *ClientEventHandler {
	return &ClientEventHandler{
		OnRead: func(r io.Reader) {
			// Set Non Blocking to use for loop
			// for {
			buf := make([]byte, 4096) // 4KB
			br, err := r.Read(buf)
			if err != nil {
				log.Println("error reading from client", err)
				return
			}
			// if br == 0 {
			// 	break
			// }
			if br > 0 {
				log.Println(string(buf[:br-1]))
			}
			// }
		},
		OnWrite: func(w io.Writer) {
			panic("Not implemented")
			// _, err := w.Write([]byte("Hello from server"))
			// if err != nil {
			// 	fmt.Println(err)
			// }
		},
		OnHangUp: func() {
			fmt.Println("Client hung up")
		},
		OnErr: func() {
			fmt.Println("Client error")
		},
	}
}

type ChatClient struct {
	sock         ClientSock
	name         string
	eventHandler *ClientEventHandler
}

func NewChatClient(cs ClientSock, name string, eventHandler *ClientEventHandler) *ChatClient {
	return &ChatClient{
		sock:         cs,
		name:         name,
		eventHandler: eventHandler,
	}
}

func (c *ChatClient) UpdateName(name string) {
	if c.name != name {
		c.name = name
	}
}
