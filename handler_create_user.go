package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mitlex/chirpy/internal/database"
)

// When the database package returns a database.User (from our CreateUser query) we will map it to this struct
// This gives us control over the JSON tags when we marshal the struct to JSON (as opposed to having the Go config option 'emit_json_tags' automatically create the tags)
type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

// handlerCreateUser takes a JSON body containing an email address from the request
// It creates a new user in the Chirpy database and if successful responds with a 201 (Created) response with the new User data in a JSON format
func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type reqParameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	reqParams := reqParameters{}
	err := decoder.Decode(&reqParams) // any missing fields in the request body JSON will have their struct values set to their zero value
	if err != nil {                   // error usually occurs due to JSON having wrong types or being invalid
		respondWithError(w, 500, "Something went wrong", err)
		return
	}

	// Create new database user with request provided email
	var newDbUser database.User
	newDbUser, err = cfg.db.CreateUser(r.Context(), reqParams.Email)
	if err != nil {
		respondWithError(w, 500, "Error creating user", err)
		return
	}

	// map new database user record to a User struct before marshalling and responding (for JSON tag control purposes - see User struct definition commentary)
	newUser := User{
		ID:        newDbUser.ID,
		CreatedAt: newDbUser.CreatedAt,
		UpdatedAt: newDbUser.UpdatedAt,
		Email:     newDbUser.Email,
	}

	// respond with HTTP Created (201) success and the new user data
	err = respondWithJSON(w, http.StatusCreated, newUser)
	if err != nil {
		respondWithError(w, 500, "Server error occurred", err)
		return
	}
}
