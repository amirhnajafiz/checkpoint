package user_data

import (
	"fmt"
	"net/http"
)

func GetData(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")

	// TODO: get user data from database

	_, _ = fmt.Fprint(w, username)
}
