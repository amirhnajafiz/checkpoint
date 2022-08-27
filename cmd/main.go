package main

import (
	"errors"
	"fmt"

	"github.com/amirhnajafiz/checkpoint/config/airbraker"
	"github.com/amirhnajafiz/checkpoint/config/server"
)

func main() {
	defer airbraker.Airbrake.Close()
	defer airbraker.Airbrake.NotifyOnPanic()

	fmt.Println("API server started ...")
	airbraker.Airbrake.Notify(errors.New("test from Airbrake"), nil)

	server.HandleRequests()
}
