// Package schema provides schema support for the database.
package schema

import (
	"context"
	_ "embed" // Embed all documents
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/pkg/errors"
)

// document represents the schema for the project.
var document = `
enum Role {
	ADMIN
	USER
}

type User {
  id: ID!
  email: String! @search(by: [hash]) @id
  name: String!
  role: Role!
  password_hash: String!
}

type Action {
	id: ID!
	name: String! @search(by: [hash]) @id
	lat: Float!
	lng: Float!
	User: User
}
`

// Schema error variables.
var (
	ErrNoSchemaExists = errors.New("no schema exitst")
	ErrInvalidSchema  = errors.New("schema doesn't match")
)

// CustomFunctions is the set of custom functions defined in the schema.
// The URL to the function is required as part of the function declaration.
type CustomFunctions struct {
	UploadFeedURL string
}

// Config contains information required for the schema document.
type Config struct {
	CustomFunctions
}

// Schema provides support for schema operations against the database
type Schema struct {
	graphql  *graphql.GraphQL
	document string
}

// New constructs a Schema value for use to manage the schema.
func New(graphql *graphql.GraphQL) (*Schema, error) {

	schema := Schema{
		graphql:  graphql,
		document: document,
	}

	return &schema, nil
}

// Create is used to create the schema in the database.
func (s *Schema) Create(ctx context.Context) error {
	query := `mutation updateGQLSchema($schema: String!) {
		updateGQLSchema(input: {
			set: { schema: $schema }
		}) {
			gqlSchema {
				schema
			}
		}
	}`
	err := s.graphql.ExecuteOnEndpoint(ctx, "admin", query, nil, graphql.WithVariable("schema", s.document))
	if err != nil {
		return errors.Wrap(err, "create schema")
	}

	return nil
}

// DropData perform an alter operatation against the configured server
// to remove all the data and schema.
func (s *Schema) DropData(ctx context.Context) error {
	query := strings.NewReader(`{"drop_op": "DATA"}`)
	if err := s.graphql.RawRequest(ctx, "alter", query, nil); err != nil {
		return errors.Wrap(err, "dropping data")
	}

	return nil
}

// DropAll perform an alter operatation against the configured server
// to remove all the data and schema.
func (s *Schema) DropAll(ctx context.Context) error {
	query := strings.NewReader(`{"drop_all": true}`)
	if err := s.graphql.RawRequest(ctx, "alter", query, nil); err != nil {
		return errors.Wrap(err, "dropping schema and data")
	}

	schema, err := s.retrieve(ctx)
	if err != nil {
		return errors.Wrap(err, "can't validate schema, db not ready")
	}

	if err := s.validate(ctx, schema); err != ErrNoSchemaExists {
		return errors.Wrap(err, "unable to drop schema and data")
	}

	return nil
}

// =============================================================================

// retrieve queries the database for the schema and handles situations
// when the database is not ready for schema operations.
func (s *Schema) retrieve(ctx context.Context) (string, error) {
	for {
		schema, err := s.query(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "Server not ready") {

				// If the context deadline exceeded then we are done trying.
				if ctx.Err() != nil {
					return "", errors.Wrap(err, "server not ready")
				}

				// We need to wait for the server to be ready for this :(.
				time.Sleep(2 * time.Second)
				continue
			}

			return "", errors.Wrap(err, "server not ready")
		}

		return schema, nil
	}
}

func (s *Schema) query(ctx context.Context) (string, error) {
	query := `query { getGQLSchema { schema }}`
	result := make(map[string]interface{})
	if err := s.graphql.ExecuteOnEndpoint(ctx, "admin", query, nil, graphql.WithVariable("result", &result)); err != nil {
		return "", errors.Wrap(err, "query schema")
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", errors.Wrap(err, "marshal schema")
	}

	return string(data), nil
}

func (s *Schema) validate(ctx context.Context, schema string) error {
	if schema == `{"getGQLSchema":null}` || schema == `{"getGQLSchema":{"schema":""}}` {
		return ErrNoSchemaExists
	}

	if len(schema) < 27 {
		return ErrInvalidSchema
	}

	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return errors.Wrap(err, "regex compile")
	}

	exp := strings.ReplaceAll(s.document, "\\n", "")
	exp = reg.ReplaceAllString(exp, "")
	schema = strings.ReplaceAll(schema[27:], "\\n", "")
	schema = strings.ReplaceAll(schema, "\\t", "")
	schema = reg.ReplaceAllString(schema, "")

	if exp != schema {
		return ErrInvalidSchema
	}

	return nil
}
