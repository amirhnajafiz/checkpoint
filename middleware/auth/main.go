package auth

import (
	"fmt"
	"github.com/amirhnajafiz/checkpoint/internal/jsonwebtoken"
	"net/http"
)

func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jsonwebtoken.ParseToken(r.Header["Token"][0])

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
