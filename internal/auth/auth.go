package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
)

// HashPassword hashes a given password
// Returns the hash and an error type
func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// CheckPasswordHash compares the password that the user entered in the HTTP request with the password stored in the database
// Returns bool (true if password matches, false otherwise) and an error type
func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, err
	}

	return match, nil
}

// MakeRefreshToken generates and returns a random 256-bit (32-byte) hex-encoded string
func MakeRefreshToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// GetAPIKey extracts the API key from the Authorization header in an HTTP request
// This comes in the format Authorization: ApiKey THE_KEY_HERE
// We strip "ApiKey " and any whitespace and return only the key string
func GetAPIKey(headers http.Header) (string, error) {
	authVal := headers.Get("Authorization")
	if authVal == "" {
		return "", errors.New("Authorization header has no value or does not exist")
	}

	// Check for malformed Authorization header
	if !strings.HasPrefix(authVal, "ApiKey ") {
		return "", errors.New("Malformed Authorization header, expecting 'ApiKey ' prefix")
	}

	// Trim off "ApiKey "
	apiKey := strings.TrimSpace(strings.TrimPrefix(authVal, "ApiKey "))

	return apiKey, nil
}
