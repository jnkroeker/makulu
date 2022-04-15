// Package user provides support for managing users in the database.
package user

import (
	"context"
	"fmt"

	"github.com/ardanlabs/graphql"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/sys/validate"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Set of error variables for CRUD operations.
var (
	ErrNotExists = errors.New("user does not exist")
	ErrExists    = errors.New("user exists")
	ErrNotFound  = errors.New("user not found")
)

// Store manages the set of APIs for user access.
type Store struct {
	log *zap.SugaredLogger
	gql *graphql.GraphQL
}

// NewStore constructs a user store for api access.
func NewStore(log *zap.SugaredLogger, gql *graphql.GraphQL) Store {
	return Store{
		log: log,
		gql: gql,
	}
}

// Add adds a new user to the database. If the user already exists
// this function will fail but the found user is returned. If the user is
// being added, the user with the id from the database is returned.
func (s Store) Add(ctx context.Context, traceID string, nu NewUser) (User, error) {
	if err := validate.Check(nu); err != nil {
		return User{}, fmt.Errorf("validating data: %w", err)
	}

	if usr, err := s.QueryByEmail(ctx, traceID, nu.Email); err == nil {
		return usr, ErrExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(nu.Password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, errors.Wrap(err, "generating password hash")
	}

	usr := User{
		Name:         nu.Name,
		Email:        nu.Email,
		Role:         nu.Role,
		PasswordHash: string(hash),
	}

	return s.add(ctx, traceID, usr)
}

// QueryByID returns the specified user from the database by the user id.
func (s Store) QueryByID(ctx context.Context, traceID string, userID string) (User, error) {
	query := fmt.Sprintf(`
query {
	getUser(id: %q) {
		id
		name
		email
		role
		password_hash
	}
}`, userID)

	s.log.Debug("%s: %s: %s", traceID, "user.QueryByID", data.Log(query))

	// the response from the call has the name of the calling function in it

	var result struct {
		GetUser User `json:"getUser"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if result.GetUser.ID == "" {
		return User{}, ErrNotFound
	}

	return result.GetUser, nil
}

// QueryByEmail returns the specified user from the database by email
func (s Store) QueryByEmail(ctx context.Context, traceID string, email string) (User, error) {
	query := fmt.Sprintf(`
query {
	queryUser(filter: { email: { eq: %q } }) {
		id
		name
		email
		role
		password_hash 
	}
}`, email)

	s.log.Debug("%s: %s: %s", traceID, "user.QueryByEmail", data.Log(query))

	// the response from the call has the name of the calling function in it
	var result struct {
		QueryUser []User `json:"queryUser"`
	}
	if err := s.gql.Execute(ctx, query, &result); err != nil {
		return User{}, errors.Wrap(err, "query failed")
	}

	if len(result.QueryUser) != 1 {
		return User{}, ErrNotFound
	}

	return result.QueryUser[0], nil
}

// =============================================================================

func (s Store) add(ctx context.Context, traceID string, usr User) (User, error) {
	var result addResult
	mutation := fmt.Sprintf(`
	mutation {
		addUser(input: [{
			name: %q 
			email: %q 
			role: %s 
			password_hash: %q 
		}])
		%s
	}`, usr.Name, usr.Email, usr.Role, usr.PasswordHash, result.document())

	s.log.Debug("%s: %s: %s", traceID, "user.Add", data.Log(mutation))

	// marshal the result of the mutation executed against the database into the result

	if err := s.gql.Execute(ctx, mutation, &result); err != nil {
		return User{}, errors.Wrap(err, "failed to add user")
	}

	if len(result.AddUser.User) != 1 {
		return User{}, errors.New("user id not returned")
	}

	usr.ID = result.AddUser.User[0].ID
	return usr, nil
}
