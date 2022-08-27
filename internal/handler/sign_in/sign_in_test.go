package sign_in

import (
	"errors"
	"github.com/amirhnajafiz/checkpoint/internal/config/airbraker"
	"testing"
)

func TestRegister(t *testing.T) {
	username := "admin-test"
	password := "super-pass"

	message := username + ":" + password

	if message != HandleRegister(username, password) {
		t.Error("sign up failed")
		airbraker.Airbrake.Notify(errors.New("sign up failed"), nil)
	}
}
