package web

import (
	"context"
	"errors"
	"time"
)

// we don't want package level variables, in general, but here
// we never want them exported and we dont want to need config for initialization

// ctxKey represents the type of value for the context key.
type ctxKey int

// key is how request values are stored/retrieved
const key ctxKey = 1

// Values represent state for each request.
//
// We're going to stick this in the context for every request.
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// GetValues returns the values from the context.
func GetValues(ctx context.Context) (*Values, error) {
	// search the context using the key
	// this will return a copy of Values indexed at the key
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return nil, errors.New("web value missing from context")
	}
	return v, nil
}

// GetTrace
func GetTraceID(ctx context.Context) string {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return "0000000000-0000-0000-0000-0000000000"
	}
	return v.TraceID
}

// SetStatusCode sets the status code back into the context.
func SetStatusCode(ctx context.Context, statusCode int) error {
	v, ok := ctx.Value(key).(*Values)
	if !ok {
		return errors.New("web value missing from context")
	}
	v.StatusCode = statusCode
	return nil
}
