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

	"github.com/jnkroeker/makulu/app/services/action-api/handlers/debug/checkgrp"
	"github.com/jnkroeker/makulu/app/services/action-api/handlers/v1/actiongrp"
	"github.com/jnkroeker/makulu/app/services/action-api/handlers/v1/testgrp"
	"github.com/jnkroeker/makulu/app/services/action-api/handlers/v1/usergrp"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/action"
	"github.com/jnkroeker/makulu/business/data/user"
	"github.com/jnkroeker/makulu/business/feeds/loader"
	"github.com/jnkroeker/makulu/business/sys/auth"
	"github.com/jnkroeker/makulu/business/web/v1/mid"
	"github.com/jnkroeker/makulu/foundation/web"
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

func DebugMux(build string, log *zap.SugaredLogger, gqlConfig data.GraphQLConfig) http.Handler {
	mux := DebugStandardLibraryMux()

	// Register debug check endpoints.
	cgh := checkgrp.Handlers{
		Build:     build,
		GqlConfig: gqlConfig,
		Log:       log,
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
	Auth   *auth.Auth
	DB     data.GraphQLConfig
	Loader loader.Config
}

// APIMux constructs an http.Handler with all application routes defined.
// Remember to always return a concrete type, never abstract for the user
func APIMux(cfg APIMuxConfig) *web.App {

	// this creates a wrapper (App) to put around our handler (tgh.Test)
	// for graceful error handling in a custom Handle() method.
	// This way the handler (tgh.Test) is not directly exposed to the mux
	app := web.NewApp(
		cfg.Shutdown,
		mid.Logger(cfg.Log),
		mid.Errors(cfg.Log),
		mid.Metrics(),
		mid.Panics(),
	)

	v1(app, cfg)

	return app
}

// v1 binds all the version 1 routes.
func v1(app *web.App, cfg APIMuxConfig) {
	const version = "v1"

	tgh := testgrp.Handlers{
		Log: cfg.Log,
	}
	app.Handle(http.MethodGet, version, "/test", tgh.Test)
	app.Handle(http.MethodGet, version, "/testauth", tgh.Test, mid.Authenticate(cfg.Auth), mid.Authorize("ADMIN"))

	// TODO: connect to Strava API using feedgrp

	// fg := feedgrp.Handlers{
	// 	Log:          cfg.Log,
	// 	GqlConfig:    cfg.DB,
	// 	LoaderConfig: cfg.Loader,
	// }
	// app.Handle(http.MethodPost, version, "/feed/upload", fg.Upload)

	act := actiongrp.Handlers{
		ActionStore: action.NewStore(
			cfg.Log,
			data.NewGraphQL(cfg.DB),
		),
	}
	app.Handle(http.MethodPost, version, "/action", act.Create, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/action/:id", act.QueryByID, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/action/user/:id", act.QueryByUser, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodPut, version, "/action/:id", act.Update, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodDelete, version, "/action/:id", act.Delete, mid.Authenticate(cfg.Auth))

	usr := usergrp.Handlers{
		UserStore: user.NewStore(
			cfg.Log,
			data.NewGraphQL(cfg.DB),
		),
		Auth: cfg.Auth,
	}
	app.Handle(http.MethodGet, version, "/users/token", usr.Token)
	app.Handle(http.MethodPost, version, "/users", usr.Create, mid.Authenticate(cfg.Auth), mid.Authorize("ADMIN"))
	app.Handle(http.MethodGet, version, "/users/:id", usr.QueryByID, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodGet, version, "/users/email/:email", usr.QueryByEmail, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodPut, version, "/users/:id", usr.Update, mid.Authenticate(cfg.Auth))
	app.Handle(http.MethodDelete, version, "/users/:id", usr.Delete, mid.Authenticate(cfg.Auth), mid.Authorize("ADMIN"))

}
