package domain

type HttpError struct {
	err  error
	code int
}

func (h HttpError) Error() string {
	return h.err.Error()
}

func (h HttpError) GetCode() int {
	return h.code
}

func (h HttpError) GetErr() error {
	return h.err
}
