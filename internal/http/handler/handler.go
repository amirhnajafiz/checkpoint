package handler

import (
	"encoding/json"
	"fmt"
	"github.com/amirhnajafiz/checkpoint/internal/jwt"
	"net/http"
	"strings"
)

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// TODO: check the user model is database

	_ = json.NewEncoder(w).Encode(HandleLogin(username, password))
}

func HandleLogin(username string, password string) map[string]string {
	response := make(map[string]string)
	token, err := jwt.GenerateToken(username + password)

	response["Token"] = token

	if err != nil {
		response["Token"] = "nil"
	}

	return response
}

func Register(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// TODO: Save to database

	_, _ = fmt.Fprint(w, HandleRegister(username, password))
}

func HandleRegister(username string, password string) string {
	return username + ":" + password
}

func GetData(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")

	// TODO: get user data from database

	_, _ = fmt.Fprint(w, HandleGetData(username))
}

func HandleGetData(username string) string {
	return strings.ToUpper(username)
}
