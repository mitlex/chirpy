package auth

import (
	"net/http"
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

func TestGetBearerToken(t *testing.T) {
	cases := []struct {
		name   string
		Header http.Header
	}{
		{
			name:   "Authorization header exists with correct form",
			Header: http.Header{"Authorization": []string{"Bearer testing"}},
		},
		{
			name:   "Authorization header exists with incorrect form",
			Header: http.Header{"Authorization": []string{"malformed header"}},
		},
		{
			name:   "Authorization header is empty",
			Header: http.Header{"Authorization": []string{""}},
		},
		{
			name:   "Authorization header doesn't exist in header",
			Header: http.Header{},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			if c.Header.Get("Authorization") == "Bearer testing" {
				tokenString, err := GetBearerToken(c.Header)
				if err != nil {
					t.Fatalf("error: could not get token: %v", err)
				}

				if tokenString != "testing" {
					t.Fatal("error: Authorization header prefix stripping failed")
				}
			} else {
				_, err := GetBearerToken(c.Header)
				if err == nil {
					t.Fatalf("error: invalid Authorization header passed test: %v", err)
				}
			}
		})
	}
}
