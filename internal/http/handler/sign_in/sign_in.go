package sign_in

import (
	"fmt"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// TODO: Save to database

	_, _ = fmt.Fprint(w, HandleRegister(username, password))
}

func HandleRegister(username string, password string) string {
	return username + ":" + password
}
