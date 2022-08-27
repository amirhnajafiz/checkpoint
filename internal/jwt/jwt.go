package jwt

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var key = []byte("mysuperseceretpharase")

// GenerateToken : We create a JWT token in GenerateToken function.
func GenerateToken(user string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	// We set the claims
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["user"] = user
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	// Then we generate the token
	tokenString, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken : We check a validation of token in this function.
func ParseToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, OK := token.Method.(*jwt.SigningMethodHMAC); !OK {
			return nil, fmt.Errorf("there was an error")
		}

		return key, nil
	})
}
