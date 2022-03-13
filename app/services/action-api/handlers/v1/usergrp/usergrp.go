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
	Auth      *auth.Auth
}

func (h Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// recall that these values were set in context when we called the app.Handle()
	// method (foundation/web/web.go) to respond to requests for creating a User.
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

// Update a user in the system
func (h Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return v1Web.NewRequestError(auth.ErrorForbidden, http.StatusForbidden)
	}

	userID := web.Param(r, "id")

	// Dont allow if a user is updating someone other than themselves
	if claims.Subject != userID {
		return v1Web.NewRequestError(auth.ErrorForbidden, http.StatusForbidden)
	}

	var usr user.User
	if err := web.Decode(r, &usr); err != nil {
		return fmt.Errorf("unable to decode payload: %w", err)
	}

	if err := h.UserStore.Update(ctx, v.TraceID, usr); err != nil {
		switch {
		case errors.Is(err, user.ErrNotExists):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s] User[%+v]: %w", userID, usr, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusOK)
}

// Delete a user from the system
func (h Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	claims, err := auth.GetClaims(ctx)
	if err != nil {
		return v1Web.NewRequestError(auth.ErrorForbidden, http.StatusForbidden)
	}

	// Can only delete an account if Admin
	if !claims.Authorized(auth.RoleAdmin) {
		return v1Web.NewRequestError(auth.ErrorForbidden, http.StatusForbidden)
	}

	userID := web.Param(r, "id")

	if err := h.UserStore.Delete(ctx, v.TraceID, userID); err != nil {
		switch {
		case errors.Is(err, user.ErrNotExists):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		default:
			return fmt.Errorf("ID[%s]: %w", userID, err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusOK)
}

// Token provides an API token for the authenticated user.
func (h Handlers) Token(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	v, err := web.GetValues(ctx)
	if err != nil {
		return web.NewShutdownError("web value missing from context")
	}

	email, pass, ok := r.BasicAuth()
	if !ok {
		err := errors.New("must provide email and password in Basic auth")
		return v1Web.NewRequestError(err, http.StatusUnauthorized)
	}

	claims, err := h.UserStore.Authenticate(ctx, v.TraceID, email, pass)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrNotFound):
			return v1Web.NewRequestError(err, http.StatusNotFound)
		case errors.Is(err, user.ErrAuthenticationFailure):
			return v1Web.NewRequestError(err, http.StatusUnauthorized)
		default:
			return fmt.Errorf("authenticating: %w", err)
		}
	}

	var tkn struct {
		Token string `json:"token"`
	}
	tkn.Token, err = h.Auth.GenerateToken(claims)
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	return web.Respond(ctx, w, tkn, http.StatusOK)
}
