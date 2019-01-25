package errors

import (
	"os"

	raven "github.com/getsentry/raven-go"
	"github.com/wearefair/k8-cross-cluster-controller/pkg/logging"
)

var sentry *raven.Client

func init() {
	var err error
	dsn := os.Getenv("SENTRY_DSN")
	env := os.Getenv("FAIR_ENV")
	if env == "" {
		env = "development"
	}
	sentry, err = raven.New(dsn)
	if err != nil {
		panic(err)
	}
	sentry.SetEnvironment(env)
}

// Reports an error to sentry
func Error(err error) error {
	if os.Getenv("ENV") == "test" {
		return err
	}

	logging.Logger.Error(err.Error())

	sentry.CaptureError(err, nil)
	return err
}
