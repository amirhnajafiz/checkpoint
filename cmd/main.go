package main

import (
	"cmd/internal/jwt"
	"fmt"
	"net/http"
)

func homePage(w http.Response, r *http.Request) {
	validation, err := jwt.GenerateToken()

	if err != nil {
		panic(err)
	}

	fmt.Println(validation)
}

func main() {
	fmt.Println("Lets go")
}
