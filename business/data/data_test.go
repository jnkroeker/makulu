package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/ardanlabs/graphql"
	"github.com/google/go-cmp/cmp"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/schema"
	"github.com/jnkroeker/makulu/business/data/user"
	"github.com/jnkroeker/makulu/business/ready"
	"github.com/jnkroeker/makulu/foundation/tests"
	"go.uber.org/zap"
)

// all tests are here instead of in the packages they are acturally
// testing because the dgraph takes a long time to spin up and be ready for testing

type TestConfig struct {
	traceID string
	log     *zap.SugaredLogger
	url     string
	schema  schema.Config
}

// TestData validates all the mutation support in data.
func TestData(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Start up dgraph in a container.
	log, url, teardown := tests.NewUnit(t)
	t.Cleanup(teardown)

	// Configure everything to run the tests.
	// Configure everything to run the tests.
	tc := TestConfig{
		traceID: "00000000-0000-0000-0000-000000000000",
		log:     log,
		url:     url,
		schema: schema.Config{
			CustomFunctions: schema.CustomFunctions{
				UploadFeedURL: "http://0.0.0.0:3000/v1/feed/upload",
			},
		},
	}

	t.Run("readiness", readiness(tc.url))
	t.Run("user", addUser(tc))
}

// waitReady provides support for making sure the database is ready to be used.
func waitReady(t *testing.T, ctx context.Context, testID int, url string) *graphql.GraphQL {
	err := data.Validate(ctx, url, time.Second)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is ready: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to to see Dgraph is ready.", tests.Success, testID)

	gqlConfig := data.GraphQLConfig{
		URL:            url,
		AuthHeaderName: "X-Travel-Auth",
		AuthToken:      "eyJhbGciOiJSUzI1NiIsImtpZCI6IjU0YmIyMTY1LTcxZTEtNDFhNi1hZjNlLTdkYTRhMGUxZTJjMSIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ3NzY5NTI2NDkuOTExNzcxLCJpYXQiOjE2MjMzNTI2NDkuOTExNzczLCJpc3MiOiJ0cmF2ZWwgcHJvamVjdCIsInN1YiI6IkFETUlOIiwiQXV0aCI6eyJST0xFIjoiQURNSU4ifX0.RjpCUPuXBI1VK49SHR0EwCRiqRTV-EITwdv8b5se__C02hbhs6-POTpjao82Ng_pIcznFzpvmevCaXCnJRJeBIOpeSCkTG3PD9ISWwGWCFA9KIaCCgTlNsUL14JzSVHrEMYOkCNDLoa0RNO_ZoHkunGyd818RwnqoxlQ6e9OCJzeKzrOgGK2_JCmBQ5G497nDEVwlP7V3zR2K0Bt21zXa2YjGgbIOj31yf6USXznlP3v9Gw8ES5sDeAd0Irp3VReQEfkn-rVDkopRLmyso-TPesDWjGTFuERAGDLifO0CiQGgCTtGZ-gPMrA3vsHAG3GsQmjBziUNmmzBErBG0J5ag",
	}
	gql := data.NewGraphQL(gqlConfig)

	schema, err := schema.New(gql)
	if err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to prepare the schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to prepare the schema.", tests.Success, testID)

	// Performing this action here breaks the current version of Dgraph.
	// To see this, uncomment this code and comment lines 96-99.
	// This code used to work on an earlier version of dgraph.
	//
	// if err := schema.DropAll(ctx); err != nil {
	// 	t.Fatalf("\t%s\tTest %d:\tShould be able to drop the data and schema: %v", tests.Failed, testID, err)
	// }
	// t.Logf("\t%s\tTest %d:\tShould be able to drop the data and schema.", tests.Success, testID)

	if err := schema.Create(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to create the schema: %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to create the schema.", tests.Success, testID)

	if err := schema.DropData(ctx); err != nil {
		t.Fatalf("\t%s\tTest %d:\tShould be able to drop the data : %v", tests.Failed, testID, err)
	}
	t.Logf("\t%s\tTest %d:\tShould be able to drop the data.", tests.Success, testID)

	return gql
}

// readiness validates the health check is working.
func readiness(url string) func(t *testing.T) {
	tf := func(t *testing.T) {
		type tableTest struct {
			name       string
			retryDelay time.Duration
			timeout    time.Duration
			success    bool
		}

		tt := []tableTest{
			{"timeout", 500 * time.Millisecond, time.Second, false},
			{"ready", 500 * time.Millisecond, 20 * time.Second, true},
		}

		t.Log("Given the need to be able to validate the database is ready.")
		{
			for testID, test := range tt {
				tf := func(t *testing.T) {
					t.Logf("\tTest %d:\tWhen waiting up to %v for the database to be ready.", testID, test.timeout)
					{
						ctx, cancel := context.WithTimeout(context.Background(), test.timeout)
						defer cancel()

						err := ready.Validate(ctx, url, test.retryDelay)
						switch test.success {
						case true:
							if err != nil {
								t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is ready: %v", tests.Failed, testID, err)
							}
							t.Logf("\t%s\tTest %d:\tShould be able to see Dgraph is ready.", tests.Success, testID)

						case false:
							if err == nil {
								t.Fatalf("\t%s\tTest %d:\tShould be able to see Dgraph is Not ready.", tests.Failed, testID)
							}
							t.Logf("\t%s\tTest %d:\tShould be able to see Dgraph is Not ready.", tests.Success, testID)
						}
					}
				}
				t.Run(test.name, tf)
			}
		}
	}
	return tf
}

// addUser validates a user node can be added to the database.
func addUser(tc TestConfig) func(t *testing.T) {
	tf := func(t *testing.T) {
		t.Log("Given the need to be able to validate storing a user")
		{
			testID := 0
			t.Logf("\tTest %d:\tWhen handling a single user.", testID)
			{
				ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
				defer cancel()

				gql := waitReady(t, ctx, testID, tc.url)

				newUser := user.NewUser{
					Name:            "Timothy Tidwell",
					Email:           "time@test.com",
					Role:            "ADMIN",
					Password:        "admin",
					PasswordConfirm: "admin",
				}

				store := user.NewStore(tc.log, gql)

				addedUser, err := store.Add(ctx, "1234", newUser)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to add a user: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to add a user", tests.Success, testID)

				retUser, err := store.QueryByID(ctx, "1235", addedUser.ID)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for a user by ID: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for a user by ID.", tests.Success, testID)

				if diff := cmp.Diff(addedUser, retUser); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff: %v", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same user.", tests.Success, testID)

				retUserTwo, err := store.QueryByEmail(ctx, "1236", addedUser.Email)
				if err != nil {
					t.Fatalf("\t%s\tTest %d:\tShould be able to query for a user by email: %v", tests.Failed, testID, err)
				}
				t.Logf("\t%s\tTest %d:\tShould be able to query for a user by email.", tests.Success, testID)

				if diff := cmp.Diff(addedUser, retUserTwo); diff != "" {
					t.Fatalf("\t%s\tTest %d:\tShould get back the same user. Diff: %v", tests.Failed, testID, diff)
				}
				t.Logf("\t%s\tTest %d:\tShould get back the same user.", tests.Success, testID)
			}
		}
	}
	return tf
}
