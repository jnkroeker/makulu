package commands

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/schema"
	"github.com/jnkroeker/makulu/business/data/user"
	"github.com/jnkroeker/makulu/business/feeds/loader"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// Schema handles the updating of the schema.
func Schema(gqlConfig data.GraphQLConfig, config schema.Config) error {
	if err := loader.UpdateSchema(gqlConfig, config); err != nil {
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

	log.Infow("Adding user: ", newUser.Name)
	if err := AddUser(log, gqlConfig, newUser); err != nil {
		if errors.Cause(err) != user.ErrExists {
			return errors.Wrap(err, "adding user")
		}
	}

	var actions = []struct {
		Name string
		Lat  float64
		Lng  float64
	}{
		{"test", 39.10984405810049, -77.56261903186942},
	}

	for _, a := range actions {
		search := loader.Search{
			Name: a.Name,
			Lat:  a.Lat,
			Lng:  a.Lng,
		}

		log.Info("adding action", a.Name)
		traceID := uuid.New().String()
		if err := loader.UpdateData(log, gqlConfig, traceID, config, search); err != nil {
			return err
		}
	}

	return nil
}
