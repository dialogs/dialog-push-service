package worker

import (
	"strconv"
)

const (
	ErrorCodeUnknown        ErrorCode = 0
	ErrorCodeUnregistered   ErrorCode = 1
	ErrorCodeBadDeviceToken ErrorCode = 2
	ErrorCodeBadRequest     ErrorCode = 3
)

type ErrorCode int

type Response struct {
	ProjectID   string
	DeviceToken string
	Error       error
}

type ResponseError struct {
	Code ErrorCode
	err  error
}

func NewResponseError(code ErrorCode, err error) *ResponseError {
	return &ResponseError{
		Code: code,
		err:  err,
	}
}

func NewResponseErrorFromAnswer(code int, err error) *ResponseError {
	return &ResponseError{
		Code: ErrorCode(code),
		err:  err,
	}
}

func NewResponseErrorBadDeviceToken(err error) *ResponseError {
	return NewResponseError(ErrorCodeBadDeviceToken, err)
}

func (r *ResponseError) Error() string {
	return strconv.Itoa(int(r.Code)) + " " + r.err.Error()
}

func (r *ResponseError) Err() error {
	return r.err
}
