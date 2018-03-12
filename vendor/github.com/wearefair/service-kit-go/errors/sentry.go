package errors

import (
	"os"

	raven "github.com/getsentry/raven-go"
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
