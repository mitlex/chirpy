package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/mitlex/chirpy/internal/database"
)

// handlerGetChirp looks for a chirp by ID in the Chirpy database and responds with it if found
func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// get path parameter value and convert to UUID
	chirpIdStr := r.PathValue("chirpID")
	chirpId, err := uuid.Parse(chirpIdStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid UUID string", err)
		return
	}

	// get chirp from database
	dbChirp, err := cfg.db.GetChirp(r.Context(), chirpId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Chirp not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Error getting chirp", err)
		return
	}

	chirp := convertDbChirpToChirpStruct(dbChirp)

	err = respondWithJSON(w, http.StatusOK, chirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}
}

// convertDbChirpToChirpStruct converts a database.Chirp to a Chirp struct with json tags
func convertDbChirpToChirpStruct(c database.Chirp) Chirp {
	return Chirp{
		ID:        c.ID,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
		Body:      c.Body,
		UserID:    c.UserID,
	}
}
