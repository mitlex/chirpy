package auth

import (
	"testing"
)

// TestPasswordHashing runs multiple tests with the HashPassword and CheckPasswordHash functions
func TestPasswordHashing(t *testing.T) {
	cases := []struct {
		name              string
		passwordToHash    string
		passwordToCompare string
		expectedMatch     bool
	}{
		{
			name:              "valid password",
			passwordToHash:    "easypassword", // imagine this is the password that is set by the original user
			passwordToCompare: "easypassword", // imagine this is the password a user enters when trying to authenticate (should work)
			expectedMatch:     true,
		},
		{
			name:              "invalid password",
			passwordToHash:    "easypassword",      // imagine this is the password that is set by the original user
			passwordToCompare: "notsoeasypassword", // imagine this is the password a user enters when trying to authenticate (should NOT work)
			expectedMatch:     false,
		},
		{
			name:              "case sensitive password",
			passwordToHash:    "easypassword",
			passwordToCompare: "EASYPASSWORD",
			expectedMatch:     false,
		},
		{
			// Verifies our auth wrapper correctly hashes and compares any string,
			// including empty input. Whether empty passwords are allowed is a
			// separate policy decision for the HTTP validation layer.
			name:              "empty password",
			passwordToHash:    "",
			passwordToCompare: "",
			expectedMatch:     true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// hash the test case password
			hash, err := HashPassword(c.passwordToHash)

			// check if HashPassword errored or generated an empty hash
			if err != nil {
				t.Fatalf("error creating hash: %v", err) // fatal ends the test - no hash = no comparison possible
			}
			if hash == "" {
				t.Fatal("no hash generated") // as above
			}

			// check if password to compare matches hash
			match, err := CheckPasswordHash(c.passwordToCompare, hash)
			if err != nil {
				t.Fatalf("error checking password against hash: %v", err) // no match result to check, end test case with fatal
			}
			if match != c.expectedMatch {
				t.Errorf("match = %v, expected %v", match, c.expectedMatch) // final assertion, so t.Errorf is sufficient; the test can finish normally
			}
		})
	}
}

// TestCheckPasswordMalformedHash tests the CheckPasswordHash function to ensure it returns an error when an invalid/malformed hash is provided to it
func TestCheckPasswordHashMalformedHash(t *testing.T) {
	_, err := CheckPasswordHash("easypassword", "IAmAMalformedInvalidHash") // the hash value is not a valid encoded Argon2id hash
	if err == nil {                                                         // therefore we do expect an error to return from CheckPasswordHash
		t.Fatal("malformed hash should cause CheckPasswordHash to return an error")
	}
}
