package middleware

import (
	"fmt"
	"net/http"

	"github.com/amirhnajafiz/checkpoint/internal/jwt"
)

func Auth(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["token"] != nil {
			token, err := jwt.ParseToken(r.Header["Token"][0])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)

				_, err = fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)

			_, _ = fmt.Fprintf(w, "Not authorized")
		}
	})
}
