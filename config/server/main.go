package server

import (
	"cmd/internal/handler/login"
	"cmd/internal/handler/sign_in"
	"cmd/internal/handler/user_data"
	"cmd/middleware/auth"
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
