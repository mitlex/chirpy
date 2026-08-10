package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mitlex/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

// handlerCreateChirp takes a chirp and user id from a JSON formatted request body, validates the chirp, and removes profanity from it
// If the chirp is valid, it is saved to the chirpy database
// Upon successful record creation, record is marshalled into a JSON object and a HTTP 201 Created response is given
func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params) // any missing fields in the request body JSON will have their struct values set to their zero value
	if err != nil {                // error usually occurs due to JSON having wrong types or being invalid
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	// check if Chirp is longer than 140 chars
	// assume that we are only handling ASCII chars (where 1 char = 1 byte)
	if len(params.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	// apply profanity filter to chirp
	params.Body = profanityFilter(params.Body)

	// save the chirp to the database
	dbChirpParams := database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserID,
	}
	var dbChirp database.Chirp
	dbChirp, err = cfg.db.CreateChirp(r.Context(), dbChirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Chirp creation error", err)
		return
	}

	// load new chirp db record into Chirp struct instance for marshalling with fixed JSON tags (see Chirp struct)
	chirp := Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}

	// respond with chirp and 201 Creation success
	err = respondWithJSON(w, http.StatusCreated, chirp)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}
}

// profanityFilter takes a chirp and replaces any of the following words with ****:
// kerfuffle, sharbert, fornax
// returns the profanity-filtered chirp
func profanityFilter(s string) string {
	clean_words := []string{}
	words := strings.Split(s, " ")
	for _, word := range words {
		originalWord := word
		lowercaseWord := strings.ToLower(word)
		if lowercaseWord == "kerfuffle" || lowercaseWord == "sharbert" || lowercaseWord == "fornax" {
			clean_words = append(clean_words, "****")
		} else {
			clean_words = append(clean_words, originalWord)
		}
	}
	result := strings.Join(clean_words, " ")
	return result
}
