package server

import (
	"github.com/amirhnajafiz/checkpoint/internal/handler/login"
	"github.com/amirhnajafiz/checkpoint/internal/handler/sign_in"
	"github.com/amirhnajafiz/checkpoint/internal/handler/user_data"
	"github.com/amirhnajafiz/checkpoint/internal/middleware/auth"
	"log"
	"net/http"
)

func HandleRequests() {
	// Auth routes
	http.HandleFunc("/api/login", login.Login)
	http.HandleFunc("/api/register", sign_in.Register)
	// Web routes
	http.Handle("/api/user", auth.IsAuthorized(user_data.GetData))

	log.Fatal(http.ListenAndServe(":5001", nil))
}
