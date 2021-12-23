package main

import (
	"cmd/internal/jwt"
	"fmt"
	"net/http"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	validation, err := jwt.GenerateToken()

	if err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
	}

	_, _ = fmt.Fprintf(w, validation)
}

func main() {
	fmt.Println("Lets go")
}
