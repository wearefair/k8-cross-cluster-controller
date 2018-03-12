package errors

import "fmt"

type baseError struct {
	code    string
	message string
}

func (e baseError) Error() string {
	message := e.message
	if e.code != "" {
		message += ", code: " + e.code
	}
	return message
}

func (e baseError) String() string {
	return e.Error()
}

func (e baseError) Code() string {
	return e.code
}

func (e baseError) Message() string {
	return e.message
}

type requestError struct {
	FairError
	statusCode int
	requestId  string
}

func (e requestError) Error() string {
	message := e.FairError.Error()
	if e.statusCode != 0 {
		message = fmt.Sprintf("%s, status code: %d", message, e.statusCode)
	}
	if e.requestId != "" {
		message += ", request id: " + e.requestId
	}
	return message
}

func (e requestError) String() string {
	return e.Error()
}

func (e requestError) StatusCode() int {
	return e.statusCode
}

func (e requestError) RequestId() string {
	return e.requestId
}
