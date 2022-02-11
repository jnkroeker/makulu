package validate

import (
	"encoding/json"
	"errors"
)

// ErrInvalidID occurs when an ID is not in a valid form.
//
// At some point we are creating our own IDs
var ErrInvalidID = errors.New("ID is not in its proper form")

// ErrorResponse is the form used for API responses from failures in the API.
//
// This project also manages a service where we have to send a response
//
// All callers to the web API will get an error string
// and if there were field model errors, then you get information about the field in error
type ErrorResponse struct {
	Error  string `json:"error"`
	Fields string `json:"fields,omitempty"`
}

// RequestError is used to pass an error during the request through the
// application with web specific context.
//
// This error type is exported so we can type assert against it outside this package
//
// this is a "trusted error" because we blindly send information back to the caller
//
// with untrusted errors (errors we don't know about) we only return statuscode
type RequestError struct {
	Err    error
	Status int
	Fields error
}

// NewRequestError wraps a provided error with an HTTP status code.
// This function shoudl be used when handlers encounter expected errors.
func NewRequestError(err error, status int) error {
	return &RequestError{err, status, nil}
}

// Error implements the error interface. It uses the default message of the wrapped error.
// This is what will be shown in the services' logs.
func (err *RequestError) Error() string {
	return err.Err.Error()
}

// FieldError is used to indicate an error with a specific request field
//
// we must ensure the all the data a user passes to us is clean,
// in order to construct the type
type FieldError struct {
	Field string `json:"field"`
	Error string `json:"error"`
}

// FieldErrors represents a collection of field errors.
type FieldErrors []FieldError

// Error implements the error interface
func (fe FieldErrors) Error() string {
	d, err := json.Marshal(fe)
	if err != nil {
		return err.Error()
	}
	return string(d)
}

// the go standard library has two functions, Is() and As()
// Is() lets you compare two error interfaces together to see if the have a common concrete value
// As() 1) lets us check if there is a concrete value of a given type inside the error
// 2) gets us a copy of the concrete error value back out

// Cause iterates through all the wrapped errors until the root
// error value is reached
//
// This allows me to avoid writing several As() functions
// in the case where the concrete error value may be one of several error types
func Cause(err error) error {
	root := err
	for {
		if err = errors.Unwrap(root); err == nil {
			return root
		}
		root = err
	}
}
