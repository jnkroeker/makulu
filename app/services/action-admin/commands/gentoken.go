package commands

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jnkroeker/makulu/business/data"
	"github.com/jnkroeker/makulu/business/data/user"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func GenToken(log *zap.SugaredLogger, gqlConfig data.GraphQLConfig, email string) error {
	if email == "" {
		fmt.Println("help: gentoken <email>")
		return ErrHelp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := user.NewStore(
		log,
		data.NewGraphQL(gqlConfig),
	)
	traceID := uuid.New().String()

	// Retrieve the user by email so we have the roles for this user.
	usr, err := store.QueryByEmail(ctx, traceID, email)
	if err != nil {
		return errors.Wrap(err, "getting user")
	}

	name := "/home/john/go/src/makulu/zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem"
	file, err := os.Open(name)
	if err != nil {
		return err
	}

	privatePem, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return err
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePem)
	if err != nil {
		return err
	}

	// ===========================================================

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := struct {
		jwt.RegisteredClaims
		Roles []string
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "phaction project",
			Subject:   usr.ID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Roles: []string{"ADMIN"},
	}

	method := jwt.GetSigningMethod("RS256")
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		return err
	}

	fmt.Println("======== TOKEN START ========")
	fmt.Println(tokenStr)
	fmt.Println("======== TOKEN END ========")

	// ==============================================================

	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the private key file.
	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return fmt.Errorf("encoding to public file: %w", err)
	}

	// =========================================================================

	fmt.Println("=========================")

	// Create the token parser to use. The algorithm used to sign the JWT must be
	// validated to avoid a critical vulnerability:
	// https://auth0.com/blog/critical-vulnerabilities-in-json-web-token-libraries/
	parser := jwt.Parser{
		ValidMethods: []string{"RS256"},
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		kid, ok := t.Header["kid"]
		if !ok {
			return nil, errors.New("missing key id (kid) in token header")
		}
		kidID, ok := kid.(string)
		if !ok {
			return nil, errors.New("user token key id (kid) must be string")
		}

		fmt.Println("KID:", kidID)

		return &privateKey.PublicKey, nil
	}

	var parsedClaims struct {
		jwt.RegisteredClaims
		Roles []string
	}

	parsedToken, err := parser.ParseWithClaims(tokenStr, &parsedClaims, keyFunc)
	if err != nil {
		return fmt.Errorf("parsing token: %w", err)
	}

	if !parsedToken.Valid {
		return errors.New("invalid token")
	}

	fmt.Println("Token Validated")

	return nil

}
