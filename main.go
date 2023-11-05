package main

import (
	"fmt"

	"github.com/dharanad/smallchat/internal"
)

func main() {
	fmt.Println("smalltalk")
	cm := &internal.ClientManager{}
	server, err := internal.NewServer(cm, 7002)
	if err != nil {
		panic(err)
	}
	if err = server.Run(); err != nil {
		panic(err)
	}
}
