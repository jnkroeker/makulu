package mid

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/jnkroeker/makulu/business/sys/metrics"
	"github.com/jnkroeker/makulu/foundation/web"
)

// This handler handles requests that have bad code that cause the program to panic
// I don't want the entire service to shutdown. I want to capture it
// and make sure the rest of the middleware runs.
// we could just let the http package handle a panic and return a 500
// but there I lose control of all my middleware

// Panics recovers from panics and converts the panic to an error so it is
// reported in Metrics and handler in Errors
func Panics() web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {

			// Defer a function the recover from a panic and set the err return
			// variable after the fact.
			//
			// this function must be defered before we define and return handler(ctx, w, r)
			// panic has to be RING ONE IN THE ONION because the panic needs to be stopped
			//
			// this defer executes in the in-between state between the function returning
			// and the calling function gaining control back
			defer func() {

				// if an error occurs with recover() how do I return the error?
				// Because I am in the in-between state, not in a state of return
				// Go has a 'named return argument' feature
				// err below is the named return
				// ** only use this pattern in situations like this
				if rec := recover(); rec != nil {

					// Stack trace will be provided
					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace))

					// Updates the metrics stored in the context
					metrics.AddPanics(ctx)
				}
			}()

			// Call the next handler and set its return value in the err variable
			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
