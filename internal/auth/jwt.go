package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// MakeJWT creates and returns a JSON Web Token (JWT)
// Uses the golang-jwt/jwt/v5 library; a golang implementation of JWTs
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	// Create the claims
	currentTime := time.Now().UTC()

	claims := &jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(currentTime),
		ExpiresAt: jwt.NewNumericDate(currentTime.Add(expiresIn)),
		Issuer:    "chirpy-access",
	}

	// Generate the JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Convert tokenSecret string to []byte
	tokenSecretKey := []byte(tokenSecret)

	// Sign the JWT: hash the header and payload with the secret to produce the signature
	signedToken, err := token.SignedString(tokenSecretKey)
	if err != nil {
		return "", err
	}

	// Return the signed JWT
	return signedToken, nil
}

// ValidateJWT validates a user provided JWT Token
// Accepts tokenString (a JWT) of format header.payload.signature and the tokenSecret held by our server
// Returns the user ID from the payload/claims
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	// Instantiate an empty RegisteredClaims struct for ParseWithClaims to fill in.
	// The payload is base64-encoded JSON, so this really is json.Unmarshal under
	// the hood: empty struct in, populated struct out.
	claims := &jwt.RegisteredClaims{}

	// Parse the JWT, verify the signature, and validate the time claims
	// (exp = expires at, nbf = not before) in a single call.
	// That's why the single err check below covers every rejection case: bad signature,
	// expired token, malformed structure, or an unexpected signing method.
	//
	// The third argument is a callback, not a value. The library decodes the
	// header first (so it knows which algorithm to use), then calls this function
	// to ask "what key should I verify with?". We ignore the token argument and
	// always hand back the same secret as []byte because that's what HMAC
	// methods require.
	//
	// We only need the claims, not the token itself, hence _
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Grab the user's ID from the token claims
	userIDStr := claims.Subject

	// convert userID string to uuid
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}
