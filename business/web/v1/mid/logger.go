package mid

import (
	"context"
	"net/http"
	"time"

	"github.com/jnkroeker/makulu/foundation/web"
	"go.uber.org/zap"
)

// How to get a logger into a web.Middleware type? (signature type Middleware func(web.Handler) web.Handler)
// DONT hide the logger in context!
// Its too much here to construct a type to hold the logger and hang this function as a method off the type
//
// Leverage closures!
// define a function that takes the logger as a parameter
// then construct a web.Middleware to return
// inside the web.Middleware construct the web.Handler that the web.Middleware must return
// use the closures created during these function constructions to use the log
func Logger(log *zap.SugaredLogger) web.Middleware {

	m := func(handler web.Handler) web.Handler {

		// we can however hide things we need for debugging in the context
		// in the foundation/web/context file, the values type will be in every http request
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing the value, request the service
			// to be shutdown gracefully.
			v, err := web.GetValues(ctx)
			if err != nil {
				return err
			}

			log.Infow("request started", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)

			// Call the next handler
			err = handler(ctx, w, r)

			log.Infow("request completed", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr,
				"statusCode", v.StatusCode, "since", time.Since(v.Now))

			// Return the error so it can be handled futher up the chain.
			return err
		}

		return h

	}

	return m
}
