// Package action provides support for managing action data in the database.
package action

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/sys/validate"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Set of error variables for CRUD operations
var (
	ErrNotFound  = errors.New("action not found")
	ErrNotExists = errors.New("user does not exist")
	ErrExists    = errors.New("user exists")
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
func (s Store) Add(ctx context.Context, traceID string, na NewAction) (Action, error) {
	if err := validate.Check(na); err != nil {
		return Action{}, fmt.Errorf("validating data: %w", err)
	}

	act := Action{
		Name: na.Name,
		Lat:  na.Lat,
		Lng:  na.Lng,
		User: na.User,
	}

	return s.add(ctx, traceID, act)
}

func (s Store) Update(ctx context.Context, traceID string, act Action) error {
	if err := validate.Check(act); err != nil {
		return fmt.Errorf("validating data: %w", err)
	}

	if _, err := s.QueryByID(ctx, traceID, act.ID); err != nil {
		return ErrNotExists
	}

	return s.update(ctx, traceID, act)
}

func (s Store) Delete(ctx context.Context, traceID, actionID string) error {
	if actionID == "" {
		return errors.New("user missing id")
	}

	if _, err := s.QueryByID(ctx, traceID, actionID); err != nil {
		return ErrNotExists
	}

	return s.delete(ctx, traceID, actionID)

}

// QueryByID returns the specified action from the database by the action id.
func (s Store) QueryByID(ctx context.Context, traceID string, actionID string) (Action, error) {
	query := fmt.Sprintf(`
query {
	getAction(id: %q) {
		id
		name
		lat
		lng
		user
	}
}`, actionID)

	s.log.Debug("%s: %s: %s", traceID, "action.QueryByID", data.Log(query))

	// the response from the call has the name of the calling function in it

	var result struct {
		GetAction Action `json:"getAction"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return Action{}, errors.Wrap(err, "query failed")
	}

	if result.GetAction.ID == "" {
		return Action{}, ErrNotFound
	}

	return result.GetAction, nil
}

// QueryByUser returns the specified action from the database by the user id.
func (s Store) QueryByUser(ctx context.Context, traceID string, userID string) (Action, error) {
	query := fmt.Sprintf(`
query {
	queryAction(filter: { User: { eq: %q } }) {
		id
		name
		lat
		lon
		user 
	}
}`, userID)

	s.log.Debug("%s: %s: %s", traceID, "action.QueryByUser", data.Log(query))

	// the response from the call has the name of the calling function in it
	var result struct {
		QueryAction []Action `json:"queryAction"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return Action{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryAction) != 1 {
		return Action{}, ErrNotFound
	}

	return result.QueryAction[0], nil
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
			user: %q
		}])
		%s
	}`, act.Name, act.Lat, act.Lng, act.User, result.document())

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

func (s Store) update(ctx context.Context, traceID string, act Action) error {
	var result updateResult
	mutation := fmt.Sprintf(`
	mutation {
		updateAction(input: {
			filter: { 
			id: [%q]
			},
			set: {
				name: %q
				lat: %f 
				lng: %f
				user: %q
			}
		})
		%s
	}
	`, act.ID, act.Name, act.Lat, act.Lng, act.User, result.document())

	s.log.Debug("%s: %s: %s", traceID, "user.Update", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to update user")
	}

	if result.UpdateAction.NumUids != 1 {
		msg := fmt.Sprintf("failed to update user: NumUids: %d", result.UpdateAction.NumUids)
		return errors.New(msg)
	}

	return nil
}

func (s Store) delete(ctx context.Context, traceID string, actionID string) error {
	var result deleteResult
	mutation := fmt.Sprintf(`
	mutation {
		deleteAction(filter: { id: [%q] })
		%s
	}`, actionID, result.document())

	s.log.Debug("%s: %s: %s", traceID, "action.Delete", data.Log(mutation))

	if err := s.gql.Execute(ctx, mutation, &result); err != nil {
		return errors.Wrap(err, "failed to delete action")
	}

	if result.DeleteAction.NumUids != 0 {
		msg := fmt.Sprintf("failed to delete action: NumUids: %d Msg: %s", result.DeleteAction.NumUids, result.DeleteAction.Msg)
		return errors.New(msg)
	}

	return nil
}
