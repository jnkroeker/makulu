// Package loader provides support for update new and old Action information
package loader

import (
	"context"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/action"
	"github.com/jnkroeker/makulu/business/data/schema"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Search represents an action and its coordinates. All fields must be
// populated for a Search to be successful.
type Search struct {
	Name string
	Lat  float64
	Lng  float64
}

// Config defines the set of mandatory settings
type Config struct {
	Filter Filter
}

// Filter represents search related refinements
type Filter struct {
	Categories []string
}

// UpdateSchema creates/updates the schema for the database.
func UpdateSchema(gqlConfig data.GraphQLConfig) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := data.Validate(ctx, gqlConfig.URL, 5*time.Second)
	if err != nil {
		return errors.Wrapf(err, "Waiting for database to be ready")
	}

	gql := data.NewGraphQL(gqlConfig)

	schema, err := schema.New(gql)
	if err != nil {
		return errors.Wrapf(err, "preparing schema")
	}

	if err := schema.Create(ctx); err != nil {
		return errors.Wrapf(err, "creating schema")
	}

	return nil
}

// UpdateData retrieves and stores the feed data for this API
func UpdateData(log *zap.SugaredLogger, gqlConfig data.GraphQLConfig, traceID string, config Config, search Search) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	gql := data.NewGraphQL(gqlConfig)
	loader := newLoader(log, gql)

	_, err := loader.upsertAction(ctx, traceID, search.Name, search.Lat, search.Lng)
	if err != nil {
		return errors.Wrapf(err, "adding action")
	}

	return nil
}

type store struct {
	action action.Store
}

type loader struct {
	log   *zap.SugaredLogger
	gql   *graphql.GraphQL
	store store
}

func newLoader(log *zap.SugaredLogger, gql *graphql.GraphQL) loader {
	return loader{
		log: log,
		gql: gql,
		store: store{
			action: action.NewStore(log, gql),
		},
	}
}

// upsertAction adds the specified action into the database
func (l loader) upsertAction(ctx context.Context, traceID string, name string, lat float64, lng float64) (action.Action, error) {
	newAction := action.Action{
		Name: name,
		Lat:  lat,
		Lng:  lng,
	}
	newAction, err := l.store.action.Add(ctx, traceID, newAction)
	if err != nil {
		return action.Action{}, errors.Wrapf(err, "adding action: %s", name)
	}

	// log.Info("feed: Work: Upserted Action: ID: %s Name: %s Lat: %f Lng: %s", newAction.ID, name, lat, lng)

	return newAction, nil
}
