package main

import (
	"cmd/internal/server"
	"fmt"
)

func main() {
	fmt.Println("API server started ...")
	server.HandleRequests()
}
