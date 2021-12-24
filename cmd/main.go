package main

import (
	"cmd/internal/jsonwebtoken"
	"cmd/middleware/auth"
	"fmt"
	"log"
	"net/http"
)

func homePage(w http.ResponseWriter, _ *http.Request) {
	validation, err := jsonwebtoken.GenerateToken()

	if err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
	}

	_, _ = fmt.Fprintf(w, validation)
}

func handleRequests() {
	http.Handle("/", auth.IsAuthorized(homePage))
	log.Fatal(http.ListenAndServe(":5001", nil))
}

func main() {
	fmt.Println("Lets go ...")
	handleRequests()
}
