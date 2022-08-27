package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amirhnajafiz/checkpoint/internal/jwt"
)

func Auth(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if value, ok := r.Header["token"]; ok {
			token, err := jwt.ParseToken(value[0])
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				_, err = fmt.Fprintf(w, err.Error())
			}

			if token.Valid {
				ctx := context.WithValue(context.Background(), "username", r.FormValue("username"))

				r.WithContext(ctx)

				endpoint(w, r)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = fmt.Fprintf(w, "Not authorized")
		}
	})
}
