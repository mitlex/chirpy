package main

import (
	"net/http"
	"sort"

	"github.com/google/uuid"
	"github.com/mitlex/chirpy/internal/database"
)

// handlerGetChirps retrieves chirps stored in the Chirpy database and responds with them in a JSON object ordered by chirpy created_at field ascending
// If the optional query parameter 'sort=desc' is provided, chirps are sorted in descending order of their created_at datetime
//
//	If 'sort=asc' the function will retain its default ascending sorting behaviour
//
// If the optional query parameter 'author_id' is provided, only chirps created by that author are returned in the response
//
//	When searching for multiple rows in a db, db drivers typically return an empty slice [] rather than sql.ErrNoRows if no matching rows exist (unlike single-row queries)
//	Returning an empty list with 200 OK is standard REST behaviour for a list endpoint when no matching items are found
func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Determine if optional 'author_id' query parameter was provided in request
	authorID := r.URL.Query().Get("author_id")

	var dbChirps []database.Chirp
	var err error

	if authorID == "" { // Get all chirps from database
		dbChirps, err = cfg.db.GetChirps(r.Context())
	} else { // Only get chirps for given author
		authorUUID, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid author_id", err)
			return
		}
		dbChirps, err = cfg.db.GetChirpsByAuthor(r.Context(), authorUUID)
	}

	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	// Sort chirps in descending order of created_at if 'sort=desc'
	// I.e. newest chirps will show first
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder == "desc" {
		sort.Slice(dbChirps, func(i, j int) bool {
			return dbChirps[i].CreatedAt.After(dbChirps[j].CreatedAt)
		})
	}

	// Load database chirps into []Chirp with json tags
	chirps := convertDbChirpsToChirpStructs(dbChirps)

	// Respond with chirps and 200 status OK
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
