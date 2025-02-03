package main

import "fmt"

func main() {

	server := NewApiServer(":3000")
	server.Run()
	fmt.Println("Hello, World!")
}