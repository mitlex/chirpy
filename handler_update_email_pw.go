package main

import (
	"encoding/json"
	"net/http"

	"github.com/mitlex/chirpy/internal/auth"
	"github.com/mitlex/chirpy/internal/database"
)

// handlerUpdateEmailAndPassword updates a users own email and password (not other users)
// Requires an access token (JWT) in the headers: Authorization: Bearer <access-token>
// Requires a new email and password in the HTTP request body
func (cfg *apiConfig) handlerUpdateEmailAndPassword(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get and validate user JWT for authentication
	userJWT, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized request", err)
		return
	}

	userId, err := auth.ValidateJWT(userJWT, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token", err)
		return
	}

	// Get the email and password from the request body
	type reqParameters struct {
		Password string `json:"password"` // as long as server uses HTTPS in prod, it's safe to send raw passwords in HTTP requests, because the entire request is encrypted
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	reqParams := reqParameters{}
	err = decoder.Decode(&reqParams) // any missing fields in the request body JSON will have their struct values set to their zero value
	if err != nil {                  // error usually occurs due to JSON having wrong types or being invalid
		respondWithError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// reject empty string passwords (we're not enforcing any password rules in this project but we will at least enforce some form of password)
	if reqParams.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password not provided", nil)
		return
	}

	// Hash the provided password before storing in database
	hashedPassword, err := auth.HashPassword(reqParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	// Update the email and hashed password in the database for the authorized user
	dbUser, err := cfg.db.UpdateUserEmailAndHashedPassword(r.Context(), database.UpdateUserEmailAndHashedPasswordParams{
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
		ID:             userId,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	resp := User{
		ID:          dbUser.ID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
		Email:       dbUser.Email,
		IsChirpyRed: dbUser.IsChirpyRed,
	}

	// Respond with updated details (except password, of course)
	err = respondWithJSON(w, http.StatusOK, resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}
}
