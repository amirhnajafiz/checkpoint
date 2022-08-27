package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/amirhnajafiz/checkpoint/internal/jwt"
)

func Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	response := make(map[string]string)
	token, err := jwt.GenerateToken(username + password)

	response["token"] = token

	if err != nil {
		response["token"] = "nil"
	}

	_ = json.NewEncoder(w).Encode(response)
}

func Register(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	_, _ = fmt.Fprint(w, username+":"+password)
}

func GetData(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	_, _ = fmt.Fprint(w, strings.ToUpper(username))
}
