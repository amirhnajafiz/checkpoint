package user_data

import (
	"cmd/config/airbraker"
	"errors"
	"testing"
)

func TestGetData(t *testing.T) {
	username := "admin"

	if "ADMIN" != HandleGetData(username) {
		t.Error("get user data failed")
		airbraker.Airbrake.Notify(errors.New("get user data failed"), nil)
	}
}
