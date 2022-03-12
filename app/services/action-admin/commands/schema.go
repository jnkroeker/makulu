package commands

import (
	"fmt"

	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/action"
	"github.com/jnkroeker/makulu/business/data/user"
	"github.com/jnkroeker/makulu/business/feeds/loader"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Schema handles the updating of the schema.
func Schema(gqlConfig data.GraphQLConfig) error {
	if err := loader.UpdateSchema(gqlConfig); err != nil {
		return err
	}

	fmt.Println("schema updated")
	return nil
}

func Seed(log *zap.SugaredLogger, gqlConfig data.GraphQLConfig, config loader.Config) error {
	// if os.Getenv("ACTION_API_KEYS_MAPS_KEY") == "" {
	// 	return errors.New("ACTION_API_KEYS_MAPS_KEY is not set with map key")
	// }

	newUser := user.NewUser{
		Name:     "John Kroeker",
		Email:    "jnkroeker@gmail.com",
		Password: "gopher",
		Role:     "ADMIN",
	}

	log.Info("Adding user: ", newUser.Name)
	id, err := AddUser(log, gqlConfig, newUser)
	if err != nil {
		if errors.Cause(err) != user.ErrExists {
			return errors.Wrap(err, "adding user")
		}
	}

	// Create action with the returned User ID
	newAction := action.NewAction{
		Name: "Stowe 01/05/22",
		Lat:  44.53005,
		Lng:  -72.78181,
		User: id,
	}

	log.Info("Adding action: ", newAction.Name)
	if err := AddAction(log, gqlConfig, newAction); err != nil {
		return err
	}

	return nil
}
