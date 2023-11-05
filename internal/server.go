package internal

import (
	"fmt"
	"log"

	syscall "golang.org/x/sys/unix"
)

// SockerBufferedReader
type SocketReader struct {
	Fd int
}

func NewSocketReader(fd int) *SocketReader {
	return &SocketReader{
		Fd: fd,
	}
}

func (s *SocketReader) Read(buf []byte) (int, error) {
	return syscall.Read(s.Fd, buf)
}

type SocketWriter struct {
	Fd int
}

func (s *SocketWriter) Write(buf []byte) (int, error) {
	return syscall.Write(s.Fd, buf)
}

func NewSocketWriter(fd int) *SocketWriter {
	return &SocketWriter{
		Fd: fd,
	}
}

type ClientManager struct {
	clients []*ChatClient
}

func (fm *ClientManager) AddClient(f *ChatClient) {
	fm.clients = append(fm.clients, f)
}

func (fm *ClientManager) RemoveClient(client *ChatClient) {
	for i, c := range fm.clients {
		if c == client {
			fm.clients = append(fm.clients[:i], fm.clients[i+1:]...)
			break
		}
	}
}

func (fm *ClientManager) Find(fd int) *ChatClient {
	for _, file := range fm.clients {
		if file.sock.Fd() == fd {
			return file
		}
	}
	return nil
}

func (fm *ClientManager) Clients() []*ChatClient {
	return fm.clients
}

type Server struct {
	sock ServerSock
	cm   *ClientManager
	pfds []syscall.PollFd
}

func NewServer(cm *ClientManager, port int) (*Server, error) {
	ss, err := CreateTcpServerSocket(port, WithEnableResuseAddress())
	if err != nil {
		return nil, err
	}
	log.Println("Created server socket")
	return &Server{
		sock: ss,
		cm:   cm,
	}, nil
}

func (s *Server) RemoveClient(c *ChatClient) {
	for i, pf := range s.pfds {
		if pf.Fd == int32(c.sock.Fd()) {
			s.pfds = append(s.pfds[:i], s.pfds[i+1:]...)
			break
		}
	}
	s.cm.RemoveClient(c)
}

func (s *Server) AddClient(c *ChatClient) {
	s.pfds = append(s.pfds, syscall.PollFd{
		Fd:     int32(c.sock.Fd()),
		Events: syscall.POLLIN | syscall.POLLHUP,
	})
	s.cm.AddClient(c)
}

func (s *Server) Run() error {
	log.Println("Server started")
	// Add server socket to poll
	s.pfds = append(s.pfds, syscall.PollFd{
		Fd:     int32(s.sock.Fd()),
		Events: syscall.POLLIN,
	})
	for {
		_, err := syscall.Poll(s.pfds, -1)
		if err != nil {
			return err
		}
		for _, pf := range s.pfds {
			if pf.Revents == 0 {
				// NoOp
				continue
			}
			if pf.Fd == int32(s.sock.Fd()) {
				if pf.Revents&syscall.POLLIN > 0 {
					// Accept new client
					cs, err := s.sock.Accept()
					if err != nil {
						continue
					}
					log.Printf("Accepted new client %d\n", cs.Fd())
					s.AddClient(NewChatClient(cs, fmt.Sprintf("user-%d", cs.Fd()), NewNoOpClientEventHandler()))
				}
			} else {
				client := s.cm.Find(int(pf.Fd))
				if client == nil {
					continue
				}
				fmt.Printf("rEvent %d\n", pf.Revents)
				if pf.Revents&syscall.POLLHUP > 0 {
					log.Printf("client %s hang up event\n", client.name)
					client.eventHandler.OnHangUp()
					syscall.Close(int(pf.Fd))
					s.RemoveClient(client)
				} else if pf.Revents&syscall.POLLERR > 0 {
					log.Printf("client %s error event\n", client.name)
					client.eventHandler.OnErr()
					syscall.Close(int(pf.Fd))
					s.RemoveClient(client)
				} else if pf.Revents&syscall.POLLIN > 0 {
					// read from FD
					log.Printf("client %s read event\n", client.name)
					client.eventHandler.OnRead(NewSocketReader(client.sock.Fd()))
				} else if pf.Revents&syscall.POLLOUT > 0 {
					log.Printf("client %s write event\n", client.name)
					// Write data to FD
					client.eventHandler.OnWrite(NewSocketWriter(client.sock.Fd()))
				}
			}
		}
	}
}
