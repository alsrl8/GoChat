package main

import (
	"GoChatServer/playground"
	"log"
)

func main() {
	err := playground.WriteTextFile()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("done")

	select {}
}
