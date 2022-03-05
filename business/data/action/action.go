// Package action provides support for managing action data in the database.
package action

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations
var (
	ErrNotFound = errors.New("action not found")
)

// Store manages the set o APIs for action access
type Store struct {
	log *zap.SugaredLogger
	gql *graphql.GraphQL
}

// NewStore constructs an action sore for api access
func NewStore(log *zap.SugaredLogger, gql *graphql.GraphQL) Store {
	return Store{
		log: log,
		gql: gql,
	}
}

// Upsert adds a new action to the database if it doesn't already exist by name
// If the action already exists, the function will return an Action value with the existing ID
func (s Store) Add(ctx context.Context, traceID string, act Action) (Action, error) {
	if act.ID != "" {
		return Action{}, errors.New("action contains id")
	}

	return s.add(ctx, traceID, act)
}

// ===================================================================

func (s Store) add(ctx context.Context, traceID string, act Action) (Action, error) {
	var result id
	mutation := fmt.Sprintf(`
	mutation {
		resp: addAction(input: [{
			name: %q
			lat: %f 
			lng: %f
		}])
		%s
	}`, act.Name, act.Lat, act.Lng, result.document())

	// s.log.Printf("%s: %s: %s", traceID, "city.Upsert", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, &result); err != nil {
		return Action{}, errors.Wrap(err, "failed to upsert action")
	}

	if len(result.Resp.Entities) != 1 {
		return Action{}, errors.New("action id not returned")
	}

	act.ID = result.Resp.Entities[0].ID
	return act, nil
}
