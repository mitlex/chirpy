package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/mitlex/chirpy/internal/auth"
)

// handlerDeleteChirp deletes a chirp from the database by its id on these conditions:
//
//	The authenticated user ID matches the user ID associated with the chirp in the database (i.e. the user is the author of the chirp)
//	The chirp exists in the database
//
// Requires an access token (JWT) in the headers: Authorization: Bearer <access-token>
func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
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

	// Get path parameter value and convert to UUID
	chirpIdStr := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(chirpIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid UUID string", err)
		return
	}

	// Get chirp from database
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error getting chirp", err)
		return
	}

	// Check if the currently authenticated user is the chirp author
	if dbChirp.UserID != userId {
		respondWithError(w, http.StatusForbidden, "Unauthorized request", err)
		return
	}

	// Currently authenticated user is the chirp author - delete the chirp
	err = cfg.db.DeleteChirp(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}

	// Respond 204 success no content
	w.WriteHeader(http.StatusNoContent)
}
