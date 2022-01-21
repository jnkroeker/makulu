// package web contains a small web framework extension
package web

import (
	"context"
	"net/http"
	"os"
	"syscall"

	"github.com/dimfeld/httptreemux/v5"
)

// App is the entrypoint into out application and what configures our context
// object for each of our http handlers.
// We need the ability, if we find any integrity issues while the service is running,
// to initiate a clean shutdown.
// Feel free to add configuration data/logic on this App struct
type App struct {
	*httptreemux.ContextMux
	shutdown chan os.Signal
}

// NewApp creates an App value that handles a set of routes for the application.
func NewApp(shutdown chan os.Signal) *App {
	return &App{
		ContextMux: httptreemux.NewContextMux(),
		shutdown:   shutdown,
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
func (a *App) Handle(method string, group string, path string, handler Handler) {

	// at the end of the day, the outermost handler must always implement the traditional
	// http.Handler interface with a function matching the ServeHTTP() signature.
	// BUT we can do anything we want inside of this function;
	// like call the custom Handler function type, passed in the last param, to handle a request
	// we can handle errors, occuring during servicing of the request, gracefully
	h := func(w http.ResponseWriter, r *http.Request) {

		// PRE CODE PROCESSING

		if err := handler(r.Context(), w, r); err != nil {
			// ERROR HANDLING
			return
		}

		// POST CODE PROCESSING
	}

	finalPath := path
	if group != "" {
		finalPath = "/" + group + path
	}

	// the only thing we can ever actually bind to the mux is using the Handle method from the mux
	// this is the true implementation of the mux; now living inside our App wrapper
	a.ContextMux.Handle(method, finalPath, h)
}
