package server

import (
	"log"
	"net/http"

	"github.com/amirhnajafiz/checkpoint/internal/http/handler"
	"github.com/amirhnajafiz/checkpoint/internal/http/middleware"
)

func HandleRequests() {
	// Auth routes
	http.HandleFunc("/api/login", handler.Login)
	http.HandleFunc("/api/register", handler.Register)
	// Web routes
	http.Handle("/api/user", middleware.Auth(handler.GetData))

	log.Fatal(http.ListenAndServe(":5001", nil))
}
