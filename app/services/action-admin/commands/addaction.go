package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/action"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// AddUser handles the creation of users.
func AddAction(log *zap.SugaredLogger, gqlConfig data.GraphQLConfig, newAction action.NewAction) error {
	if newAction.Name == "" || newAction.Lat == 0 || newAction.Lng == 0 || newAction.User == "" {
		fmt.Printf("help: addaction %s %f %f %s", newAction.Name, newAction.Lat, newAction.Lng, newAction.User)
		return ErrHelp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := action.NewStore(
		log,
		data.NewGraphQL(gqlConfig),
	)
	traceID := uuid.New().String()

	act, err := store.Add(ctx, traceID, newAction)
	if err != nil {
		return errors.Wrap(err, "adding action")
	}

	fmt.Println("action id:", act.ID)
	return nil
}
