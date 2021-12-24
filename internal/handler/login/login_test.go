package login

import (
	"cmd/config/airbraker"
	"errors"
	"testing"
)

func TestLogin(t *testing.T) {
	username := "admin"
	password := "super-pass"

	response := HandleLogin(username, password)

	if response["Token"] == "nil" {
		t.Error("login failed")
		airbraker.Airbrake.Notify(errors.New("login failed"), nil)
	}
}
