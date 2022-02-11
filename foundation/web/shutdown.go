package web

// This represents an error type that we are making foundational

import (
	"errors"
)

// shutdownError is a type used to help with the graceful termination of the service.
//
// prefer custom error type to be unexported, this way people can't type assert against them.
// IsShutdown() is a support function to determine the type of an error
//
// if they are exported, a method set is nice or an interface with a bunch of error types
// that can use the same method set
type shutdownError struct {
	Message string
}

// NewShutdownError returns an error that causes the framework to signal
// a graceful shutdown.
func NewShutdownError(message string) error {
	return &shutdownError{message}
}

// Error is the implementation of the error interface.
// Why pointer semantics
func (se *shutdownError) Error() string {
	return se.Message
}

// IsShutdown checks to see if the shutdown error is contained
// in the specified error value.
// If the error is a shutdownError, its marshalled into the se variable
// eliminates type checking
//
// check out the net package for functions that return bool like this for error types
func IsShutdown(err error) bool {
	var se *shutdownError
	return errors.As(err, &se)
}
