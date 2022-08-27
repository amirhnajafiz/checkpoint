package airbraker

import "github.com/airbrake/gobrake/v5"

func New() *gobrake.Notifier {
	return gobrake.NewNotifierWithOptions(&gobrake.NotifierOptions{
		ProjectId:   384477,
		ProjectKey:  "1914b401317b91a2192d0f899c8ad943",
		Environment: "production",
	})
}
