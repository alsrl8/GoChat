package main

import (
	"GoChatServer/internal/playground"
	"log"
)

func main() {
	err := playground.WriteTextFile()
	if err != nil {
		log.Println(err)
		panic(err)
	}

	playground.WaitForInterruptSignal()
}
