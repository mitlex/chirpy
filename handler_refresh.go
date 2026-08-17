package main

import (
	"net/http"
	"time"

	"github.com/mitlex/chirpy/internal/auth"
)

// handlerRefresh responds with a new JWT if a valid refresh token exists for the given user.
// It does not accept a request body but does require a refresh token in the headers: Authorization: Bearer <refresh-token>
func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get the refresh token from the HTTP request header
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error retrieving token from request header", err)
		return
	}

	// Check for refresh token match in database and get associated user
	userId, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	// Refresh token is valid - mint new JWT
	userJwt, err := auth.MakeJWT(userId, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}

	type response struct {
		Token string `json:"token"`
	}
	resp := response{
		Token: userJwt,
	}

	// Respond with fresh JWT
	err = respondWithJSON(w, http.StatusOK, resp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}
}
