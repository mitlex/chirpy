package auth

import (
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
