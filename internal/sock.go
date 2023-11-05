package internal

import syscall "golang.org/x/sys/unix"

/*
MacOS Header Files Location:
/Library/Developer/CommandLineTools/SDKs/MacOSX14.sdk/usr/include
*/

type ClientSock interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
	Close() error
	Fd() int
}

type ServerSock interface {
	Accept() (ClientSock, error)
	listen(int) error
	bind(int) error
	Close() error
	Fd() int
}

// Note: Sockets are actually implemented as file descriptors
// so all sockets are file descriptors but the converse is not true
// Sockets are the uniform interface between the user process and the network protocol stacks in the kernel
type Socket struct {
	fd int
}

// Sock Constructor
func NewSock(fd int) *Socket {
	return &Socket{
		fd: fd,
	}
}

// Fd returns the file descriptor of the socket
func (s *Socket) Fd() int {
	return s.fd
}

// Close closes the file descriptor
func (s *Socket) Close() error {
	return syscall.Close(s.fd)
}

func (s *Socket) Read(buf []byte) (int, error) {
	return syscall.Read(s.fd, buf)
}

func (s *Socket) Write(buf []byte) (int, error) {
	return syscall.Write(s.fd, buf)
}

// Accept extracts the first connection request on the queue of pending connections
func (s *Socket) Accept() (ClientSock, error) {
	// Ref: https://man7.org/linux/man-pages/man2/accept.2.html
	// accept client connection on server socket
	cfd, _, err := syscall.Accept(s.fd)
	if err != nil {
		return nil, err
	}
	return NewSock(cfd), nil
}

// Listen marks the socket referred to by fd as a passive socket
func (s *Socket) listen(backlog int) error {
	// Ref: https://man7.org/linux/man-pages/man2/listen.2.html
	// Use socket to accept incoming connection
	return syscall.Listen(s.fd, backlog)
}

// Bind associates a socket with a port on the local machine
func (s *Socket) bind(port int) error {
	// Network byte order is just big endian
	sa := syscall.SockaddrInet4{
		Port: port,
		Addr: [4]byte{0, 0, 0, 0},
	}
	// Ref: https://man7.org/linux/man-pages/man2/bind.2.html
	// Bind access to TCP socket
	return syscall.Bind(s.Fd(), &sa)
}

// CreateTcpSocket creates a TCP socket
func CreateTcpSocket() (*Socket, error) {
	// Ref: https://man7.org/linux/man-pages/man2/socket.2.html
	// Since there is only a single protocol for AF_INET and SOCK_STREAM which is TCP
	// we set proto to 0
	fd, err := syscall.Socket(syscall.AF_INET /* Internet Protocol */, syscall.SOCK_STREAM /* stream type */, 0)
	if err != nil {
		return nil, err
	}
	return NewSock(fd), nil
}

type SockOpts func(*Socket) error

func WithEnableResuseAddress() SockOpts {
	return func(s *Socket) error {
		// Ref: https://man7.org/linux/man-pages/man3/setsockopt.3p.html
		// Set options at socket level to reuse address in case the server restarts
		return syscall.SetsockoptInt(s.Fd(), syscall.SOL_SOCKET /*protocol level*/, syscall.SO_REUSEADDR, 1)
	}
}

// CreateTcpServerSocket creates a TCP server socket
func CreateTcpServerSocket(port int, opts ...SockOpts) (ServerSock, error) {
	serverSock, err := CreateTcpSocket()
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		err = opt(serverSock)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	err = serverSock.bind(port)
	if err != nil {
		return nil, err
	}
	err = serverSock.listen(syscall.SOMAXCONN)
	if err != nil {
		return nil, err
	}
	return serverSock, nil
}
