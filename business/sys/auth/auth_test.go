package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/jnkroeker/makulu/business/sys/auth"
	"github.com/jnkroeker/makulu/foundation/tests"
)

func TestAuth(t *testing.T) {
	t.Log("Given the need to be able to authenticate and authorize access.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single user,", testID)
		{
			const keyID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"
			privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create a private key: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create a private key.", tests.Failed, testID)

			a, err := auth.New(keyID, &keyStore{pk: privateKey})
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create an authenticator: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create an authenticator.", tests.Success, testID)

			claims := auth.Claims{
				RegisteredClaims: jwt.RegisteredClaims{
					Issuer:    "action project",
					Subject:   "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1",
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(8760 * time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
				Roles: []string{auth.RoleAdmin},
			}

			token, err := a.GenerateToken(claims)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a JWT: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a JWT.", tests.Success, testID)

			parsedClaims, err := a.ValidateToken(token)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse the claims: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse the claims.", tests.Success, testID)

			if exp, got := len(claims.Roles), len(parsedClaims.Roles); exp != got {
				t.Logf("\t\tTest %d:\texp: %v", testID, exp)
				t.Logf("\t\tTest %d:\texp: %v", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the expected number of roles: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the expected number of roles.", tests.Success, testID)

			if exp, got := claims.Roles[0], parsedClaims.Roles[0]; exp != got {
				t.Logf("\t\tTest %d:\texp: %v", testID, exp)
				t.Logf("\t\tTest %d:\texp: %v", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the expected roles: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the expected roles.", tests.Success, testID)
		}

	}
}

// ==============================================================

type keyStore struct {
	pk *rsa.PrivateKey
}

func (ks *keyStore) PrivateKey(kid string) (*rsa.PrivateKey, error) {
	return ks.pk, nil
}

func (ks *keyStore) PublicKey(kid string) (*rsa.PublicKey, error) {
	return &ks.pk.PublicKey, nil
}
