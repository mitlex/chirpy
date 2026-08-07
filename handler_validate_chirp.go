package main

import (
	"encoding/json"
	"net/http"
	"strings"
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

	// apply profanity filter to chirp
	params.Body = profanityFilter(params.Body)

	// respond with chirp
	err = respondWithJSON(w, 200, map[string]string{"cleaned_body": params.Body})
	if err != nil {
		respondWithError(w, 500, "Server error occurred", err)
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
