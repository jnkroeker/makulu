package testgrp

import (
	"context"
	"math/rand"
	"net/http"

	"github.com/jnkroeker/makulu/foundation/web"
	"go.uber.org/zap"
)

type Handlers struct {
	Log *zap.SugaredLogger
}

func (h Handlers) Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if n := rand.Intn(100); n%2 == 0 {
		//return errors.New("untrusted error")
		//return validate.NewRequestError(errors.New("trusted error"), http.StatusBadRequest)
		//panic("testing panic")
	}

	status := struct {
		Status string
	}{
		Status: "OK",
	}

	// we don't want handler developers leaving encoding up to interpretation
	// we ensure consistency of API response by abstracting that
	// `return json.NewEncoder(w).Encode(status)`
	return web.Respond(ctx, w, status, http.StatusOK)
}
