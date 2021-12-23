package main

import (
	"cmd/internal/jwt"
	"fmt"
	jwt2 "github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
)

var mySigningKey = []byte("mysuperseceretpharase")

func homePage(w http.ResponseWriter, r *http.Request) {
	validation, err := jwt.GenerateToken()

	if err != nil {
		_, _ = fmt.Fprintf(w, err.Error())
	}

	_, _ = fmt.Fprintf(w, validation)
}

func isAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header["Token"] != nil {
			token, err := jwt2.Parse(r.Header["Token"][0], func(token *jwt2.Token) (interface{}, error) {
				if _, OK := token.Method.(*jwt2.SigningMethodHMAC); !OK {
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

func handleRequests() {
	http.Handle("/", isAuthorized(homePage))
	log.Fatal(http.ListenAndServe(":5001", nil))
}

func main() {
	fmt.Println("Lets go ...")
	handleRequests()
}
