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

// AddUser handles the creation of users.
func AddUser(log *zap.SugaredLogger, gqlConfig data.GraphQLConfig, newUser user.NewUser) (string, error) {
	if newUser.Name == "" || newUser.Email == "" || newUser.Password == "" || newUser.Role == "" {
		fmt.Println("help: adduser <name> <email> <password> <role>")
		return "", ErrHelp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := user.NewStore(
		log,
		data.NewGraphQL(gqlConfig),
	)
	traceID := uuid.New().String()

	usr, err := store.Add(ctx, traceID, newUser)
	if err != nil {
		return "", errors.Wrap(err, "adding user")
	}

	fmt.Println("user id:", usr.ID)
	return usr.ID, nil
}
