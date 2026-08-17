package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mitlex/chirpy/internal/auth"
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
		Token string `json:"token"`
	}

	defer r.Body.Close()
	type reqParameters struct {
		Password         string `json:"password"` // as long as server uses HTTPS in prod, it's safe to send raw passwords in HTTP requests, because the entire request is encrypted
		Email            string `json:"email"`
		ExpiresInSeconds int    `json:"expires_in_seconds"` // not time.Duration since we want to accept times in seconds, not nanoseconds, we'll handle time.Duration manipulation with the provided integer
	}

	decoder := json.NewDecoder(r.Body)
	reqParams := reqParameters{}
	err := decoder.Decode(&reqParams) // any missing fields in the request body JSON will have their struct values set to their zero value
	if err != nil {                   // error usually occurs due to JSON having wrong types or being invalid
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	// Determine JWT expiration: defaults to 1hr if unset, maximum 1hr
	expiresIn := time.Duration(reqParams.ExpiresInSeconds) * time.Second
	if expiresIn <= 0 || expiresIn > time.Hour {
		expiresIn = time.Hour
	}

	// get user from database by provided email address
	dbUser, err := cfg.db.GetUserByEmail(r.Context(), reqParams.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err) // do NOT leak that the email exists - good security practice
		return
	}

	// check password matches hashed password for this user
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

	// create JWT for user and assign it to response
	userJwt, err := auth.MakeJWT(loggedInUser.ID, cfg.jwtSecret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}

	resp := response{
		User:  loggedInUser,
		Token: userJwt,
	}

	err = respondWithJSON(w, http.StatusOK, resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}
}
