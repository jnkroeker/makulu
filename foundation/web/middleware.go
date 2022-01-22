package web

// Middleware is a function designed to run some code before and/or after
// another Handler. It is designed to remove boilerplate or other concerns not
// direct to any given Handler
type Middleware func(Handler) Handler

// wrapMiddlware creates a new handler by wrapping middleware around a final
// handler.
// It constructs the onion from inside out, the outer layer always being the mux handler.
// The middlewares' Handlers will be executed by requests in the order
// they are provided.
func wrapMiddleware(mw []Middleware, handler Handler) Handler {

	// Loop backwards through the middleware, invoking each. Replace the handler
	// with the new wrapped handler. Looping backwards ensures that the first
	// middleware of the slice is the first to be executed by requests.
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}
