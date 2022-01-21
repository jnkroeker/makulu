// Package handlers contains the full set of handler functions and routes
// supported by the web api.
package handlers

// when importing the net/http/pprof package, its init() function is called.
// The init() function binds the exported functions to the default default server mux.
// We can't use the default server mux in production, because it is unclear what other endpoints are bound to it.
// We add the endpoints to a new mux in this package.
import (
	"expvar"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/dimfeld/httptreemux/v5"
	"github.com/jnkroeker/makulu/app/services/action-api/handlers/debug/checkgrp"
	"github.com/jnkroeker/makulu/app/services/action-api/handlers/v1/testgrp"
	"go.uber.org/zap"
)

// DebugStandardLibraryMux registers all the debug routes from the standard library
// into a new mux bypassing the use of the DefaultServerMux.
// Using the DefaultServerMux would be a security risk since a dependency could inject a
// handler into our service without us knowing it.
func DebugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.Handle("/debug/vars", expvar.Handler())

	return mux
}

func DebugMux(build string, log *zap.SugaredLogger) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register debug check endpoints.
	cgh := checkgrp.Handlers{
		Build: build,
		Log:   log,
	}
	mux.HandleFunc("/debug/readiness", cgh.Readiness)
	mux.HandleFunc("/debug/liveness", cgh.Liveness)

	return mux
}

// APIMuxConfig contains all the mandatory systems required by handlers.
type APIMuxConfig struct {
	Shutdown chan os.Signal
	Log      *zap.SugaredLogger
	// Metrics  *metrics.Metrics
	// Auth     *auth.Auth
	// DB       *sqlx.DB
}

// APIMux constructs an http.Handler with all application routes defined.
// *always return a concrete type
func APIMux(cfg APIMuxConfig) *httptreemux.ContextMux {
	mux := httptreemux.NewContextMux()

	tgh := testgrp.Handlers{
		Log: cfg.Log,
	}
	mux.Handle(http.MethodGet, "/v1/test", tgh.Test)

	return mux
}
