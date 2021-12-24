package main

import (
	"cmd/config/airbraker"
	"cmd/config/server"
	"errors"
	"fmt"
)

func main() {
	defer airbraker.Airbrake.Close()
	defer airbraker.Airbrake.NotifyOnPanic()

	fmt.Println("API server started ...")
	airbraker.Airbrake.Notify(errors.New("test from Airbrake"), nil)

	server.HandleRequests()
}
