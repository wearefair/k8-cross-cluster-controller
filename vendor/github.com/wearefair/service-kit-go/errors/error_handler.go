package errors

import (
	"os"
	"runtime"

	"go.uber.org/zap"
	"golang.org/x/net/context"

	"github.com/wearefair/service-kit-go/logging"
)

// Context value keys are compared by type, golang suggests that
// we shouldn't just use a string key
// https://golang.org/pkg/context/#WithValue
type fairErrorTagsKey string

const (
	tagsKey = fairErrorTagsKey("fair-tags")
)

// WithTags builds a new context from the current context with the provided tags.
// Tags are set on the new context's metadata on error reports and application traces
func WithTags(ctx context.Context, tags map[string]string) context.Context {
	currTags, ok := ctx.Value(tagsKey).([]map[string]string)
	if !ok {
		currTags = []map[string]string{}
	}
	currTags = append(currTags, tags)
	return context.WithValue(ctx, tagsKey, currTags)
}

// Reports an error to sentry
func Error(ctx context.Context, err error) error {
	return ErrorWithTags(ctx, err, nil)
}

// Adds additional context to the error at the error location
func ErrorWithTags(ctx context.Context, err error, tags map[string]string) error {
	if os.Getenv("ENV") == "test" {
		return err
	}

	log := logging.WithRequestId(ctx).With(zap.Error(err))
	// If we are logging a FairError, set the attributes
	// on the logger.
	if fairErr, ok := err.(FairError); ok {
		if code := fairErr.Code(); code != "" {
			log = log.With(zap.String("errorCode", code))
		}
		if reqErr, ok := fairErr.(RequestError); ok {
			if statusCode := reqErr.StatusCode(); statusCode != 0 {
				log = log.With(zap.Int("status", statusCode))
			}
			if requestId := reqErr.RequestId(); requestId != "" {
				log = log.With(zap.String("requestId", requestId))
			}
		}
		log.Error(fairErr.Message())
	} else {
		// Otherwise log the full error text.
		log.Error(err.Error())
	}
	if tags == nil {
		tags = extractTagsFromContext(ctx)
	} else {
		tags = mapMerge(extractTagsFromContext(ctx), tags)
	}
	sentry.CaptureError(err, tags)
	return err
}

// Blocks until all errors have been reported.
// Useful if you want to exit the program after reporting an exception
func Flush() {
	if os.Getenv("ENV") != "test" {
		sentry.Wait()
	}
}

// Like Error, but blocks until the error has been reported
// and then calls os.Exit(1) unless running with ENV=test, in which case
// it calls panic instead.
func Fatal(ctx context.Context, err error) {
	_ = Error(ctx, err)
	Flush()
	if os.Getenv("ENV") != "test" {
		os.Exit(1)
	} else {
		panic(err)
	}
}

// Recovers from a panic (except for a runtime.Error) and reports the panic to
// sentry
func DontPanic(ctx context.Context) error {
	recovered := recover()
	if recovered != nil {
		switch err := recovered.(type) {
		case runtime.Error:
			_ = Error(ctx, err)
			// so long and thanks for all the fish!
			panic(err)
		case error:
			_ = Error(ctx, err)
			return err
		}
	}
	return nil
}

func extractTagsFromContext(ctx context.Context) map[string]string {
	newTags := map[string]string{}
	tags, ok := ctx.Value(tagsKey).([]map[string]string)
	if !ok {
		return map[string]string{}
	}
	for _, tag := range tags {
		newTags = mapMerge(newTags, tag)
	}
	return newTags
}

func mapMerge(dest map[string]string, src map[string]string) map[string]string {
	for key, val := range src {
		dest[key] = val
	}
	return dest
}
