package errors

import "fmt"

type RequestError struct {
	err        error
	statusCode int
	rawURL     string
}

func (r *RequestError) StatusCode() int {
	return r.statusCode
}

func (r *RequestError) RawURL() string {
	return r.rawURL
}

func (r *RequestError) Error() string {
	return r.err.Error()
}

func (r *RequestError) Unwrap() error {
	return r.err
}

func NewRequestError(err error) *RequestError {
	return &RequestError{err: err}
}

func NewRequestErrorf(format string, args ...any) *RequestError {
	return &RequestError{err: fmt.Errorf(format, args...)}
}

func (r *RequestError) WithStatusCode(statusCode int) *RequestError {
	r.statusCode = statusCode
	return r
}

func (r *RequestError) WithRawURL(rawURL string) *RequestError {
	r.rawURL = rawURL
	return r
}
