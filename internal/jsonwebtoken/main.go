package jsonwebtoken

import (
	"github.com/dgrijalva/jwt-go"
	"time"
)

// This will be our key
var key = []byte("mysuperseceretpharase")

// GenerateToken : We create a JWT token in GenerateToken function
func GenerateToken() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	// We set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["user"] = "Robert Rood"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	// Then we generate the token
	tokenString, err := token.SignedString(key)

	if err != nil {
		panic(err.Error())
		return "", err
	}

	return tokenString, nil
}
