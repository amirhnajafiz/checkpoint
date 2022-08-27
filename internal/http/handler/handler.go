package handler

import (
	"fmt"
	"net/http"

	"github.com/airbrake/gobrake/v5"
	"github.com/amirhnajafiz/checkpoint/internal/jwt"
)

type Handler struct {
	Air     *gobrake.Notifier
	Storage map[string]string
}

func (h *Handler) LoginUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if _, ok := h.Storage[username]; !ok {
		w.WriteHeader(http.StatusNotFound)

		_, _ = fmt.Fprint(w, "user not found")

		return
	}

	if password != h.Storage[username] {
		w.WriteHeader(http.StatusUnauthorized)

		_, _ = fmt.Fprint(w, "password does not match")

		return
	}

	token, err := jwt.GenerateToken(username + password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		_, _ = fmt.Fprint(w, err.Error())

		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintf(w, "token:\n%s", token)
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	h.Storage[username] = password
	h.Air.Notify("new user", nil)

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetUserData(w http.ResponseWriter, r *http.Request) {
	username := r.Context().Value("username").(string)

	w.WriteHeader(http.StatusOK)

	_, _ = fmt.Fprint(w, h.Storage[username])
}
