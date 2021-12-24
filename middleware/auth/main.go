package auth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"net/http"
)

var mySigningKey = []byte("mysuperseceretpharase")

func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, OK := token.Method.(*jwt.SigningMethodHMAC); !OK {
					return nil, fmt.Errorf("there was an error")
				}
				return mySigningKey, nil
			})

			if err != nil {
				_, _ = fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {
			_, _ = fmt.Fprintf(w, "Not authorized")
		}
	})
}
