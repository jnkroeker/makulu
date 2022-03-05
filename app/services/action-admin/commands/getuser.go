package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/user"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// GetUser returns information about a user by email
func GetUser(log *zap.SugaredLogger, gqlConfig data.GraphQLConfig, email string) error {
	if email == "" {
		fmt.Println("help: getuser <email>")
		return ErrHelp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := user.NewStore(
		log,
		data.NewGraphQL(gqlConfig),
	)
	traceID := uuid.New().String()

	usr, err := store.QueryByEmail(ctx, traceID, email)
	if err != nil {
		return errors.Wrap(err, "getting user")
	}

	fmt.Printf("user: %#v\n", usr)
	return nil
}
