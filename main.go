package main

import (
	"fmt"
	"log"
)

func main() {
	store, err := NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}
	if err := store.Init(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Store is ready")

	server := NewAPIServer(":8080", store)

	server.Run()
}
