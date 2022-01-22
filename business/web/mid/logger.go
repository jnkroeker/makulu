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

		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// important to know the different log lines assocated with different requests
			traceId := "000000000000"
			statusCode := http.StatusOK
			now := time.Now()

			log.Infow("request started", "traceid", traceId, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)

			err := handler(ctx, w, r)

			log.Infow("request completed", "traceid", traceId, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr,
				"statusCode", statusCode, "since", time.Since(now))

			return err
		}

		return h

	}

	return m
}
