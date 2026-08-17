package main

import (
	"net/http"

	"github.com/mitlex/chirpy/internal/auth"
)

// handlerRevoke revokes a refresh token sent as part of the request headers
// It does not accept a request body but does require a refresh token in the headers: Authorization: Bearer <refresh-token>
func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Get the refresh token from the HTTP request header
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Error retrieving token from request header", err)
		return
	}

	// Revoke the refresh token
	err = cfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}

	// Respond 204 and no body - respondWithJSON and a nil payload will still technically return a JSON body and violates the NoContent status code
	w.WriteHeader(http.StatusNoContent)
}
