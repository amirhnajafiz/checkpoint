package server

import (
	"cmd/internal/handler"
	"log"
	"net/http"
)

func HandleRequests() {
	http.HandleFunc("/api/login", handler.Login)
	http.HandleFunc("/api/register", handler.Register)
	log.Fatal(http.ListenAndServe(":5001", nil))
}
