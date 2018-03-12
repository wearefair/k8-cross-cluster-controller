package errors

import "net/http"

type FairError interface {
	error

	Code() string

	Message() string
}

func New(code, message string) FairError {
	return &baseError{
		code:    code,
		message: message,
	}
}

type RequestError interface {
	FairError

	StatusCode() int

	RequestId() string
}

func NewRequestError(err FairError, resp *http.Response) RequestError {
	return &requestError{
		FairError:  err,
		statusCode: resp.StatusCode,
		requestId:  resp.Header.Get("x-fair-request-id"),
	}
}

func NewManualRequestError(err FairError, statusCode int, requestId string) RequestError {
	return &requestError{
		FairError:  err,
		statusCode: statusCode,
		requestId:  requestId,
	}
}
