package main

import (
	"GoChatServer/internal/server"
)

func main() {
	err := server.NewChatServer(":8080").Run()
	if err != nil {
		panic(err)
	}
}
