package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWT(t *testing.T) {
	cases := []struct {
		name              string
		makeJwtSecret     string
		validateJwtSecret string
		expiresIn         time.Duration
		userID            uuid.UUID
	}{
		{
			name:              "valid jwt",
			makeJwtSecret:     "notSoSecret",
			validateJwtSecret: "notSoSecret",
			expiresIn:         20 * time.Second,
		},
		{
			name:              "invalid jwt incorrect secret",
			makeJwtSecret:     "notSoSecret",
			validateJwtSecret: "superSecret",
			expiresIn:         20 * time.Second,
		},
		{
			name:              "invalid jwt expired jwt",
			makeJwtSecret:     "notSoSecret",
			validateJwtSecret: "notSoSecret",
			expiresIn:         -1,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// make the JWT
			token, err := MakeJWT(c.userID, c.makeJwtSecret, c.expiresIn)
			if err != nil {
				t.Fatalf("error creating JWT: %v", err)
			}

			// valid case
			if c.makeJwtSecret == c.validateJwtSecret && c.expiresIn > 0 { // equal secrets and not expired
				_, err = ValidateJWT(token, c.validateJwtSecret)
				if err != nil {
					t.Fatalf("error: JWT invalid: %v", err)
				}
			}

			// invalid cases
			if c.makeJwtSecret == c.validateJwtSecret && c.expiresIn < 0 { // expired token
				_, err = ValidateJWT(token, c.validateJwtSecret)
				if err == nil {
					t.Fatalf("error: JWT valid past expiry: %v", err)
				}
			}

			if c.makeJwtSecret != c.validateJwtSecret { // mis-matched secrets
				_, err = ValidateJWT(token, c.validateJwtSecret)
				if err == nil {
					t.Fatalf("error: JWT valid with mis-matching secrets: %v", err)
				}
			}
		})
	}
}
