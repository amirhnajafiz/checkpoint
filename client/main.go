package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

var key = []byte("mysuperseceretpharase")

func generateToken() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["user"] = "Robert Rood"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(key)

	if err != nil {
		panic(err.Error())
		return "", err
	}

	return tokenString, nil
}

func main() {
	fmt.Println(generateToken())
}
