package usergrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jnkroeker/makulu/business/data/user"
	"github.com/jnkroeker/makulu/business/sys/auth"
	"github.com/jnkroeker/makulu/business/sys/validate"
	v1Web "github.com/jnkroeker/makulu/business/web/v1"
	"github.com/jnkroeker/makulu/foundation/web"
)

// Handlers manages the set of user endpoints
type Handlers struct {
	UserStore user.Store
	// Auth *auth.Auth
}

func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// recall that these values were set in context we called the app.Handle()
	// method (foundation/web/web.go) to respond to requests for creating a user.
	// Every time the mux receives a request that is passed to this method,
	// new values are set in the context
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	var nu user.NewUser
	if err := web.Decode(r, &nu); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	usr, err := h.UserStore.Add(ctx, v.TraceID, nu)
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

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return v1Web.NewRequestError(auth.ErrorForbidden, http.StatusForbidden)
	}

	userID := web.Param(r, "id")

	// If you are not an admin and looking to retrieve someone other than yourself
	if !claims.Authorized(auth.RoleAdmin) && claims.Subject != userID {
		return v1Web.NewRequestError(auth.ErrorForbidden, http.StatusForbidden)
	}

	usr, err := h.UserStore.QueryByID(ctx, v.TraceID, userID)
	if err != nil {
		switch {
		case errors.Is(err, validate.ErrInvalidID):
			return v1Web.NewRequestError(err, http.StatusBadRequest)
		case errors.Is(err, user.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", userID, err)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}

func (h Handlers) QueryByEmail(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	email := web.Param(r, "email")

	usr, err := h.UserStore.QueryByEmail(ctx, v.TraceID, email)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("email[%s]: %w", email, err)
		}
	}

	return web.Respond(ctx, w, usr, http.StatusOK)
}
