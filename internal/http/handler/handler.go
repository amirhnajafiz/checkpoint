package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/amirhnajafiz/checkpoint/internal/jwt"
)

func LoginUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	token, err := jwt.GenerateToken(username + password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		_, _ = fmt.Fprint(w, err.Error())
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "token:\n%s", token)
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, username+":"+password)
}

func GetUserData(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, strings.ToUpper(username))
}
