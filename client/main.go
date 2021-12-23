package main

import (
	"cmd/internal/jsonwebtoken"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	token, err := jsonwebtoken.GenerateToken()

	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:5001/", nil)
	req.Header.Set("Token", token)
	res, _ := client.Do(req)

	body, _ := ioutil.ReadAll(res.Body)

	fmt.Println(body)
}
