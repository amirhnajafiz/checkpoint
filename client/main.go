package main

import (
	"cmd/server/internal/jwt"
	"fmt"
)

func main() {
	token, err := jwt.GenerateToken()

	if err != nil {
		panic(err)
	}

	fmt.Println(token)
}
