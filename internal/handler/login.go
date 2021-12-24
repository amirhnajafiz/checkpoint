package handler

import (
	"cmd/internal/jsonwebtoken"
	"encoding/json"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// TODO: check the user model is database

	response := make(map[string]string)
	token, err := jsonwebtoken.GenerateToken(username + password)

	response["Token"] = token

	if err != nil {
		response["Token"] = "nil"
	}

	_ = json.NewEncoder(w).Encode(response)
}
