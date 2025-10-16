package main

import "GoChatServer/internal/database"

func main() {
	_, err := database.ConnectToMongo()
	if err != nil {
		panic(err)
	}

	_, err = database.ConnectToPostgres()
	if err != nil {
		panic(err)
	}
}
