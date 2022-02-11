// Package metrics constructs the metrics the application will track.
package metrics

import (
	"context"
	"expvar"
)

// This holds the single instance of the metrics value needed for
// collecting metrics. The EXPVAR package is already based on a singleton
// for the different metrics that are registered with the package so there
// isn't much choice.
var m *metrics

// =======================================================================

// Metrics represents the set of metrics we gather. These fields are
// safe to be accessed concurrently thanks to expvar. Now extra abstraction required.
type metrics struct {
	goroutines *expvar.Int
	requests   *expvar.Int
	errors     *expvar.Int
	panics     *expvar.Int
}

// init constructs the metrics value that will be used to capture metrics.
// the metrics value is stored in a package level variable since everything
// inside of expvar is registered as a singleton. The use of once will make
// sure this initialization only happens once.
func init() {
	m = &metrics{
		goroutines: expvar.NewInt("goroutines"),
		requests:   expvar.NewInt("requests"),
		errors:     expvar.NewInt("errors"),
		panics:     expvar.NewInt("panics"),
	}
}

// =======================================================================

// Metics will be supported through use of CONTEXT

// ctxKeyMetric represents the type of value for the context key.
type ctxKey int

// key is how metric values are stored/retrieved
const key ctxKey = 1

// =======================================================================

// Define an API

// the api keeps all metrics contained in this package

// Add more of these functions when a metric needs to be collected in
// different parts of the codebase. This will keep this package the
// central authority for metrics and metrics won't get lost.

// Set sets the metrics data into the context.
func Set(ctx context.Context) context.Context {
	return context.WithValue(ctx, key, m)
}

// AddGoroutines increments the goroutines metric by 1.
func AddGoroutines(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		// for every hundred requests, add goroutine
		if v.requests.Value()%100 == 0 {
			v.goroutines.Add(1)
		}
	}
}

// AddGoroutines increments the goroutines metric by 1.
func AddRequests(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.requests.Add(1)
	}
}

// AddGoroutines increments the goroutines metric by 1.
func AddErrors(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.errors.Add(1)
	}
}

// AddGoroutines increments the goroutines metric by 1.
func AddPanics(ctx context.Context) {
	if v, ok := ctx.Value(key).(*metrics); ok {
		v.panics.Add(1)
	}
}
