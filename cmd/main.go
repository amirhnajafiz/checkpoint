package main

import (
	"cmd/internal/jwt"
	"fmt"
	"log"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	validation, err := jwt.GenerateToken()

	if err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
	}

	_, _ = fmt.Fprintf(w, validation)
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	log.Fatal(http.ListenAndServe(":5001", nil))
}

func main() {
	fmt.Println("Lets go ...")
	handleRequests()
}
