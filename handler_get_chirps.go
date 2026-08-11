package main

import (
	"net/http"

	"github.com/mitlex/chirpy/internal/database"
)

// handlerGetChirps retrieves all chirps stored in the Chirpy database and responds with them in a JSON object ordered by chirpy created_at field ascending
func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// get chirps from database
	dbChirps, err := cfg.db.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error getting chirps", err)
		return
	}

	// load database chirps into []Chirp with json tags
	chirps := convertDbChirpsToChirpStructs(dbChirps)

	// respond with chirps and 200 status OK
	err = respondWithJSON(w, http.StatusOK, chirps)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}
}

// convertDbChirpsToChirpStructs converts a slice of database.Chirp structs to a slice of Chirp structs with json tags
//
// Note on pre-allocating slice capacity:
// pre-allocating the new slice capacity to match the size of the old slice has multiple benefits compared to instantiating an empty slice with 0 or few elements ([]Chirp{}):
// 1. Go allocates a single array with exact capacity needed from the start
// 2. No capacity reallocations occur during the loop
// 3. No copying of existing elements occurs as the slice grows
// 4. Less pressure on garbage collector
//   - every reallocation leaves behind an abandoned array that needs garbage collected
//   - pre-allocating the array of required size removes all that garbage collection
//
// All of this results in less CPU cycles and lower latency
func convertDbChirpsToChirpStructs(dbChirps []database.Chirp) []Chirp {
	chirps := make([]Chirp, len(dbChirps))
	for i, dbChirp := range dbChirps {
		chirps[i] = convertDbChirpToChirpStruct(dbChirp)
	}
	return chirps
}
