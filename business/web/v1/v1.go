// Package v1 represents types used by the web application for v1.
package v1

type RequestError struct {
	Err    error
	Status int
}

// Wraps a provided error with an HTTP status code.
// This function shoudl be used when handlers encounter unexpected errors.
func NewRequestError(err error, status int) error {
	return &RequestError{err, status}
}

// Error implements the error interface. It uses the default
// message of the wrapped error. This is what shows in the service logs.
func (err *RequestError) Error() string {
	return err.Err.Error()
}
