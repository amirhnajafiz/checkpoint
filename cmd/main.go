package main

import (
	"cmd/internal/server"
	"errors"
	"fmt"
	"github.com/airbrake/gobrake/v5"
)

var airbrake = gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
	ProjectId:   384477,
	ProjectKey:  "1914b401317b91a2192d0f899c8ad943",
	Environment: "production",
})

func main() {
	defer airbrake.Close()
	defer airbrake.NotifyOnPanic()
	fmt.Println("API server started ...")
	airbrake.Notify(errors.New("test from Airbrake"), nil)
	server.HandleRequests()
}
