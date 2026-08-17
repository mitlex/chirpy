package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mitlex/chirpy/internal/auth"
	"github.com/mitlex/chirpy/internal/database"
)

// handlerLoginUser takes a JSON body containing an email address and password from the HTTP request
// It looks up the user by their email address and checks the given password matches the stored hashed password for that user
// If the given password matches the hashed password stored for that user email, returns HTTP status OK and a copy of their user resource (excluding the hashed password)
func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
	// response embeds User rather than declaring a named field (i.e. `User User`).
	// Embedding promotes User's fields to the outer struct, so encoding/json marshals
	// them at the top level alongside Token, producing a flat object:
	//   {"id":..., "created_at":..., "updated_at":..., "email":..., "token":...}
	// A named field would instead nest them under a "User" key, which is not the
	// shape the API contract specifies.
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	defer r.Body.Close()
	type reqParameters struct {
		Password string `json:"password"` // as long as server uses HTTPS in prod, it's safe to send raw passwords in HTTP requests, because the entire request is encrypted
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	reqParams := reqParameters{}
	err := decoder.Decode(&reqParams) // any missing fields in the request body JSON will have their struct values set to their zero value
	if err != nil {                   // error usually occurs due to JSON having wrong types or being invalid
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	// Get user from database by provided email address
	dbUser, err := cfg.db.GetUserByEmail(r.Context(), reqParams.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err) // do NOT leak that the email exists - good security practice
		return
	}

	// Check password matches hashed password for this user
	match, err := auth.CheckPasswordHash(reqParams.Password, dbUser.HashedPassword)
	if err != nil || !match {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	loggedInUser := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
		// Omit the hashed password in response for security purposes
	}

	// Create user JWT
	userJwt, err := auth.MakeJWT(loggedInUser.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}

	// Store refresh token in chirpy database
	userRefreshTokenDb, err := cfg.db.CreateRefreshToken(r.Context(),
		database.CreateRefreshTokenParams{
			Token:     auth.MakeRefreshToken(),
			UserID:    dbUser.ID,
			ExpiresAt: time.Now().UTC().Add(60 * 24 * time.Hour), // 60 days
		})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}

	// Create response payload
	resp := response{
		User:         loggedInUser,
		Token:        userJwt,
		RefreshToken: userRefreshTokenDb.Token,
	}

	err = respondWithJSON(w, http.StatusOK, resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}
}
