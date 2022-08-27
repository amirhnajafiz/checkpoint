package cmd

import (
	"log"
	"net/http"

	"github.com/amirhnajafiz/checkpoint/internal/airbraker"
	"github.com/amirhnajafiz/checkpoint/internal/http/handler"
	"github.com/amirhnajafiz/checkpoint/internal/http/middleware"
)

func Execute() {
	air := airbraker.New()

	defer air.Close()
	defer air.NotifyOnPanic()

	air.Notify("api server started", nil)

	http.HandleFunc("/api/login", handler.LoginUser)
	http.HandleFunc("/api/register", handler.RegisterUser)
	http.Handle("/api/user", middleware.Auth(handler.GetUserData))

	if err := http.ListenAndServe(":5001", nil); err != nil {
		log.Printf("error: %v\n", err)
	}
}
