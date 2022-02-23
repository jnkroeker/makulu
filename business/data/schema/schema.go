// Package schema provides schema support for the database.
package schema

import (
	"bytes"
	"context"
	_ "embed" // Embed all documents
	"html/template"

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
  email: String!
  name: String!
  role: Role!
  password_hash: String!
  date_created: DateTime!
  date_updated: DateTime!
}

type Action {
	id: ID!
	name: String!
	lat: Float!
	lng: Float!
	User: String!
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
func New(graphql *graphql.GraphQL, config Config) (*Schema, error) {

	// Create the final schema document with the variable replacements by
	// processing the template.
	tmpl := template.New("schema")
	if _, err := tmpl.Parse(document); err != nil {
		return nil, errors.Wrap(err, "parsing template")
	}
	var doc bytes.Buffer
	vars := map[string]interface{}{
		"UploadFeedURL": config.CustomFunctions.UploadFeedURL,
	}
	if err := tmpl.Execute(&doc, vars); err != nil {
		return nil, errors.Wrap(err, "executing template")
	}

	schema := Schema{
		graphql:  graphql,
		document: doc.String(),
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
