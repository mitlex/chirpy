package main

import (
	"encoding/json"
	"net/http"

	"github.com/mitlex/chirpy/internal/auth"
)

// handlerLoginUser takes a JSON body containing an email address and password from the HTTP request
// It looks up the user by their email address and checks the given password matches the stored hashed password for that user
// If the given password matches the hashed password stored for that user email, returns HTTP status OK and a copy of their user resource (excluding the hashed password)
func (cfg *apiConfig) handlerLoginUser(w http.ResponseWriter, r *http.Request) {
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

	// get user from database by provided email address
	dbUser, err := cfg.db.GetUserByEmail(r.Context(), reqParams.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", err)
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

	err = respondWithJSON(w, http.StatusOK, loggedInUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}
}
