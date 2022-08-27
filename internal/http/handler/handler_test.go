package handler

import (
	"errors"
	"testing"

	"github.com/amirhnajafiz/checkpoint/internal/config/airbraker"
	"github.com/amirhnajafiz/checkpoint/internal/jwt"
)

func TestLogin(t *testing.T) {
	username := "admin"
	password := "super-pass"

	if _, err := jwt.GenerateToken(username + password); err != nil {
		t.Error("login failed")

		airbraker.Airbrake.Notify(errors.New("login failed"), nil)
	}
}

func TestRegister(t *testing.T) {
	username := "admin-test"
	password := "super-pass"

	message := username + ":" + password

	if message != HandleRegister(username, password) {
		t.Error("sign up failed")
		airbraker.Airbrake.Notify(errors.New("sign up failed"), nil)
	}
}

func TestGetData(t *testing.T) {
	username := "admin"

	if "ADMIN" != HandleGetData(username) {
		t.Error("get user data failed")
		airbraker.Airbrake.Notify(errors.New("get user data failed"), nil)
	}
}
