package user_data

import (
	"fmt"
	"net/http"
	"strings"
)

func GetData(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")

	// TODO: get user data from database

	_, _ = fmt.Fprint(w, HandleGetData(username))
}

func HandleGetData(username string) string {
	return strings.ToUpper(username)
}
