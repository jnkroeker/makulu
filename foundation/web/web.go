// package web contains a small web framework extension
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/google/uuid"
)

// App is the entrypoint into out application and what configures our context
// object for each of our http handlers.
// We need the ability, if we find any integrity issues while the service is running,
// to initiate a clean shutdown.
// Feel free to add configuration data/logic on this App struct
type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
	mw       []Middleware
}

// NewApp creates an App value that handles a set of routes for the application.
// The `mw` parameter allows passing ZERO to many middleware functions.
// We dont want to the slice; in that case we must pass nil in cases we dont need middleware
// APIs that require to pass nil are not as accurate as the could be.
func NewApp(shutdown chan os.Signal, mw ...Middleware) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
		mw:         mw,
	}
}

// SignalShutdown is use to gracefully shutdown the app when an integrity issue is identified
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// A Handler is a type that handles an http request within our mini framework
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// Handle sets a handler function for a given HTTP method and path pair
//
// This is overriding the ContextMux Handle method with our own implementation
//
// There are middlewares that need to be applied at a handler level; like authentication
func (a *App) Handle(method string, group string, path string, handler Handler, mw ...Middleware) {

	// First wrap middleware specific to the passed in handler function.
	handler = wrapMiddleware(mw, handler)

	// Add the application's general middleware to the handler chain.
	handler = wrapMiddleware(a.mw, handler)

	// at the end of the day, the outermost handler must always implement the traditional
	// http.Handler interface with a function matching the ServeHTTP() signature.
	// BUT we can do anything we want inside of this function;
	// like call the custom Handler function type, passed in the last param, to handle a request
	// we can add middleware, handle errors, occuring during servicing of the request, gracefully
	h := func(w http.ResponseWriter, r *http.Request) {

		// We would like to log that 'Logging Started' right here, but
		// we cant log here though because this is a foundational layer library.
		// We need a way to inject code from the business layer here (aka middleware)
		// We handle this will the `wrapMiddleware` functions above
		// Visually you can think of each layer of middleware being called before calling the next middleware
		// and wrapping the next handler as this comment does

		// Pull the context from the request
		ctx := r.Context()

		// Set the context with the required values
		// process the request
		//
		// could, later, associate a userid with a traceid to help with debugging ;)
		v := Values{
			TraceID: uuid.New().String(),
			Now:     time.Now(),
		}

		ctx = context.WithValue(ctx, key, &v)

		// handler is now one function, wrapped with middleware, that will call all the middleware functions
		if err := handler(ctx, w, r); err != nil {
			// Logging error - handle it
			// We need a way to inject code from the business layer here (aka middleware)
			a.SignalShutdown()
			return
		}

		// We would like to log that 'Logging Ended' right here, but
		// we cant log here though because this is a foundational layer library.
		// We need a way to inject code from the business layer here (aka middleware)
		// We handle this will the `wrapMiddleware` functions above
		// Visually you can think of each layer of middleware being called before calling the next middleware
		// and wrapping the next handler as this comment does
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	// the only thing we can ever actually bind to the mux is using the Handle method from the mux
	// this is the true implementation of the mux; now living inside our App wrapper
	a.ContextMux.Handle(method, finalPath, h)
}
