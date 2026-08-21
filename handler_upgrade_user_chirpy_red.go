package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/mitlex/chirpy/internal/auth"
)

// handlerUpgradeUserToChirpyRed marks a Chirpy user as a Chirpy Red member when Polka (our pretend payment provider) sends us a webhook saying the user subscribed to Chirpy Red
func (cfg *apiConfig) handlerUpgradeUserToChirpyRed(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type reqParameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	decoder := json.NewDecoder(r.Body)
	reqParams := reqParameters{}
	err := decoder.Decode(&reqParams) // any missing fields in the request body JSON will have their struct values set to their zero value
	if err != nil {                   // error usually occurs due to JSON having wrong types or being invalid
		respondWithError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Ensure the webhook came from Polka by checking API key stored in Authorization header
	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "unauthorized request", err)
		return
	}
	if apiKey != cfg.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "unauthorized request", err)
		return
	}

	// Check if event is user.upgraded and respond with 204 if not (we don't care about any other events)
	if reqParams.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Event is "user.updated" - update user in database to mark them as a Chirpy Red member
	_, err = cfg.db.UpgradeUserToChirpyRed(r.Context(), reqParams.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "User not found", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	// respond to Polka with a 204
	// Hypothetically, if Polka existed, they would use this response code to know the webhook was received successfully
	// If the response code is anything other than 2XX, they would retry the request
	w.WriteHeader(http.StatusNoContent)
}
