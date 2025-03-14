package main

import (
	"log"
)

func main() {
	store, err := NewPostgresStorage()
	if err != nil {
        log.Fatal(err)
    }

	if err := store.Init(); 
	err != nil {
        log.Fatal(err)
    }

	server := NewApiServer(":3000", store)
	server.Run()
}