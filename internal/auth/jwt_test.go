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
		name          string
		header        http.Header
		expectedToken string
		expectError   bool
	}{
		{
			name:          "Valid token",
			header:        http.Header{"Authorization": []string{"Bearer testing"}},
			expectedToken: "testing",
			expectError:   false,
		},
		{
			name:          "Valid token with extra whitespace",
			header:        http.Header{"Authorization": []string{"Bearer   testing  "}},
			expectedToken: "testing",
			expectError:   false,
		},
		{
			name:          "Malformed header without Bearer prefix",
			header:        http.Header{"Authorization": []string{"malformed header"}},
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "Empty header string",
			header:        http.Header{"Authorization": []string{""}},
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "Missing Authorization header",
			header:        http.Header{},
			expectedToken: "",
			expectError:   true,
		},
		{
			name:          "Wrong scheme (ApiKey instead of Bearer)",
			header:        http.Header{"Authorization": []string{"ApiKey some-token"}},
			expectedToken: "",
			expectError:   true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			token, err := GetBearerToken(c.header)

			if c.expectError {
				if err == nil {
					t.Fatalf("expected an error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if token != c.expectedToken {
				t.Fatalf("expected token %q, got %q", c.expectedToken, token)
			}
		})
	}
}
