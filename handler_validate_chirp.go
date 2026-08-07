package main

import (
	"encoding/json"
	"net/http"
)

// handlerValidateChirpEndpoint
func handlerValidateChirpEndpoint(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params) // any missing fields in the request body JSON will have their struct values set to their zero value
	if err != nil {                // error usually occurs due to JSON having wrong types or being invalid
		respondWithError(w, 500, "Something went wrong", err)
		return
	}

	// check if Chirp is longer than 140 chars
	// assume that we are only handling ASCII chars (where 1 char = 1 byte)
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long", nil)
		return
	}

	// Chirp is valid
	err = respondWithJSON(w, 200, map[string]bool{"valid": true})
	if err != nil {
		respondWithError(w, 500, "Server error occurred", err)
		return
	}
}
