package main

// import (
// 	"fmt"
// 	"log"

// 	syscall "golang.org/x/sys/unix"
// )

// func main_single_thread() {
// 	// Note: Sockets are actually implemented as file descriptors
// 	// so all sockets are file descriptors but the converse is not true
// 	serverSock, err := createTcpServerSocket(7515)
// 	if err != nil {
// 		panic(err)
// 	}
// 	for {
// 		// Ref: https://man7.org/linux/man-pages/man2/accept.2.html
// 		// accept the client connect
// 		clientFd, _, err := syscall.Accept(serverSock)
// 		if err != nil {
// 			log.Println("error", err)
// 			continue
// 		}
// 		fmt.Printf("connected to client. fd: %d\n", clientFd)
// 		go func(fd int) {
// 			for {
// 				buf := make([]byte, 256)
// 				rn, err := syscall.Read(fd, buf)
// 				if err != nil {
// 					syscall.Close(fd)
// 					break
// 				}
// 				// Dont write entire buffer
// 				// Write same number of character which were read
// 				wn, err := syscall.Write(fd, buf[:rn])
// 				if err != nil {
// 					syscall.Close(fd)
// 					break
// 				}
// 				if rn != wn {
// 					fmt.Printf("something went wrong. rn %d wn %d\n", rn, wn)
// 				}

// 			}
// 			fmt.Println("client closed")
// 		}(clientFd)
// 	}
// }
