package user_data

import (
	"errors"
	"github.com/amirhnajafiz/checkpoint/internal/config/airbraker"
	"testing"
)

func TestGetData(t *testing.T) {
	username := "admin"

	if "ADMIN" != HandleGetData(username) {
		t.Error("get user data failed")
		airbraker.Airbrake.Notify(errors.New("get user data failed"), nil)
	}
}
