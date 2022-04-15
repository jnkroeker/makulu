// Package user provides a core business API.
// This layer is for adding business logic around data/store access
package user

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/jnkroeker/makulu/business/data/user"
	"go.uber.org/zap"
)

type Core struct {
	log  *zap.SugaredLogger
	user user.Store
}

func NewCore(log *zap.SugaredLogger, gql *graphql.GraphQL) Core {
	return Core{
		log:  log,
		user: user.NewStore(log, gql),
	}
}

func (c Core) Create(ctx context.Context, nu user.NewUser) (user.User, error) {

	// PERFORM PRE BUSINESS OPERATIONS

	usr, err := c.user.Add(ctx, "1234", nu)
	if err != nil {
		return user.User{}, fmt.Errorf("create: %w", err)
	}

	// PERFORM POST BUSINESS OPERATIONS

	return usr, nil
}
