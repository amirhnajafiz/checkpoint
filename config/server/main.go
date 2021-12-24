package server

import (
	"cmd/internal/handler"
	"cmd/middleware/auth"
	"log"
	"net/http"
)

func HandleRequests() {
	// Auth routes
	http.HandleFunc("/api/login", handler.Login)
	http.HandleFunc("/api/register", handler.Register)
	// Web routes
	http.Handle("/api/user", auth.IsAuthorized(handler.GetData))

	log.Fatal(http.ListenAndServe(":5001", nil))
}
