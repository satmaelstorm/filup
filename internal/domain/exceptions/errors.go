package exceptions

import "errors"

type ApiError struct {
	err  error
	code int
}

func (a ApiError) Error() string {
	return a.err.Error()
}

func (a ApiError) GetCode() int {
	return a.code
}

func (a ApiError) GetErr() error {
	return a.err
}

func NewApiError(code int, msg string) ApiError {
	return ApiError{
		err:  errors.New(msg),
		code: code,
	}
}
