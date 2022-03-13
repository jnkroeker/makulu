package actiongrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jnkroeker/makulu/business/data/action"
	"github.com/jnkroeker/makulu/business/sys/validate"
	v1Web "github.com/jnkroeker/makulu/business/web/v1"
	"github.com/jnkroeker/makulu/foundation/web"
)

// Handlers manages the set of user endpoints
type Handlers struct {
	ActionStore action.Store
	// Auth *auth.Auth
}

func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// recall that these values were set in context when we called the app.Handle()
	// method (foundation/web/web.go) to respond to requests for creating an Action.
	// Every time the mux receives a request that is passed to this method,
	// new values are set in the context
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var act action.NewAction
	if err := web.Decode(r, &act); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	usr, err := h.ActionStore.Add(ctx, v.TraceID, act)
	if err != nil {
		return fmt.Errorf("user[%+v]: %w", &usr, err)
	}

	return web.Respond(ctx, w, usr, http.StatusCreated)
}

func (h Handlers) QueryByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	actionID := web.Param(r, "id")

	usr, err := h.ActionStore.QueryByID(ctx, v.TraceID, actionID)
	if err != nil {
		switch {
		case errors.Is(err, validate.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, action.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", actionID, err)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (h Handlers) QueryByUser(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	user := web.Param(r, "user")

	usr, err := h.ActionStore.QueryByUser(ctx, v.TraceID, user)
	if err != nil {
		switch {
		case errors.Is(err, action.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("email[%s]: %w", user, err)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}
