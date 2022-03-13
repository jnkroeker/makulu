package main

import (
	"expvar"
	"fmt"
	"os"

	"github.com/ardanlabs/conf"
	"github.com/jnkroeker/makulu/app/services/action-admin/commands"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/user"
	"github.com/jnkroeker/makulu/business/feeds/loader"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// construct the application logger
	log, err := initLogger("ACTION-ADMIN")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := run(log); err != nil {
		if errors.Cause(err) != commands.ErrHelp {
			log.Errorw("startup", "ERROR", err)
		}
		os.Exit(1)
	}
}

func run(log *zap.SugaredLogger) error {

	// =============================================
	// Configuration

	var cfg struct {
		conf.Version
		Args   conf.Args
		Dgraph struct {
			URL            string `conf:"default:http://0.0.0.0:8080"`
			AuthHeaderName string `conf:"default:X-Action-Auth"`
			AuthToken      string
		}
		Search struct {
			Categories []string `conf:"default:cycling;skiing;crossfit"`
			// Radius     int      `conf:"default:5000"`
		}
	}
	cfg.Version.SVN = build
	cfg.Version.Desc = "copyright information here"

	const prefix = "ACTION"
	if err := conf.Parse(os.Args[1:], prefix, &cfg); err != nil {
		switch err {
		case conf.ErrHelpWanted:
			usage, err := conf.Usage(prefix, &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		case conf.ErrVersionWanted:
			version, err := conf.VersionString(prefix, &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(version)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// For convenience with the training material, an ADMIN token is provided
	if cfg.Dgraph.AuthToken == "" {
		cfg.Dgraph.AuthToken = "eyJhbGciOiJSUzI1NiIsImtpZCI6IjU0YmIyMTY1LTcxZTEtNDFhNi1hZjNlLTdkYTRhMGUxZTJjMSIsInR5cCI6IkpXVCJ9.eyJleHAiOjQ3NzY5NTI2NDkuOTExNzcxLCJpYXQiOjE2MjMzNTI2NDkuOTExNzczLCJpc3MiOiJ0cmF2ZWwgcHJvamVjdCIsInN1YiI6IkFETUlOIiwiQXV0aCI6eyJST0xFIjoiQURNSU4ifX0.RjpCUPuXBI1VK49SHR0EwCRiqRTV-EITwdv8b5se__C02hbhs6-POTpjao82Ng_pIcznFzpvmevCaXCnJRJeBIOpeSCkTG3PD9ISWwGWCFA9KIaCCgTlNsUL14JzSVHrEMYOkCNDLoa0RNO_ZoHkunGyd818RwnqoxlQ6e9OCJzeKzrOgGK2_JCmBQ5G497nDEVwlP7V3zR2K0Bt21zXa2YjGgbIOj31yf6USXznlP3v9Gw8ES5sDeAd0Irp3VReQEfkn-rVDkopRLmyso-TPesDWjGTFuERAGDLifO0CiQGgCTtGZ-gPMrA3vsHAG3GsQmjBziUNmmzBErBG0J5ag"
	}

	// =============================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars
	expvar.NewString("build").Set(build)
	log.Infow("starting admin service", "version", build)
	defer log.Infow("admin service shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Infow("startup", "config", out)

	// =============================================
	// Commands

	gqlConfig := data.GraphQLConfig{
		URL:            cfg.Dgraph.URL,
		AuthHeaderName: cfg.Dgraph.AuthHeaderName,
		AuthToken:      cfg.Dgraph.AuthToken,
	}

	switch cfg.Args.Num(0) {
	case "schema":

		if err := commands.Schema(gqlConfig); err != nil {
			return errors.Wrap(err, "updating schema")
		}

	case "seed":
		config := loader.Config{
			Filter: loader.Filter{
				Categories: cfg.Search.Categories,
			},
		}

		if err := commands.Seed(log, gqlConfig, config); err != nil {
			return errors.Wrap(err, "seeding database")
		}
	case "adduser":
		newUser := user.NewUser{
			Name:     cfg.Args.Num(1),
			Email:    cfg.Args.Num(2),
			Password: cfg.Args.Num(3),
			Role:     cfg.Args.Num(4),
		}

		// TODO: do something with the created user id
		if _, err := commands.AddUser(log, gqlConfig, newUser); err != nil {
			return errors.Wrap(err, "adding user")
		}
	case "getuser":
		email := cfg.Args.Num(1)
		if err := commands.GetUser(log, gqlConfig, email); err != nil {
			return errors.Wrap(err, "getting user")
		}
	case "keygen":
	case "gentoken":
		email := cfg.Args.Num(1)
		if err := commands.GenToken(log, gqlConfig, email); err != nil {
			return errors.Wrap(err, "generating token")
		}
	default:

	}

	return nil
}

func initLogger(service string) (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.DisableStacktrace = true
	config.InitialFields = map[string]interface{}{
		"service": service,
	}

	// build the logger
	log, err := config.Build()
	if err != nil {
		fmt.Println("Error constructing logger:", err)
		os.Exit(1)
	}

	return log.Sugar(), nil
}
