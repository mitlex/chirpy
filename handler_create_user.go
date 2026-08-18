package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mitlex/chirpy/internal/auth"
	"github.com/mitlex/chirpy/internal/database"
)

// When the database package returns a database.User (from our CreateUser query) we will map it to this struct
// This gives us control over the JSON tags when we marshal the struct to JSON (as opposed to having the Go config option 'emit_json_tags' automatically create the tags)
type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

// handlerCreateUser takes a JSON body containing an email address and password from the HTTP request
// It creates a new user in the Chirpy database and if successful responds with a 201 (Created) response with the new User data in a JSON format
func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	type reqParameters struct {
		Password string `json:"password"` // as long as server uses HTTPS in prod, it's safe to send raw passwords in HTTP requests, because the entire request is encrypted
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	reqParams := reqParameters{}
	err := decoder.Decode(&reqParams) // any missing fields in the request body JSON will have their struct values set to their zero value
	if err != nil {                   // error usually occurs due to JSON having wrong types or being invalid
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	// reject empty string passwords (we're not enforcing any password rules in this project but we will at least enforce some form of password)
	if reqParams.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Password not provided", nil)
		return
	}

	// Hash the provided password before storing in database
	hashedPassword, err := auth.HashPassword(reqParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong", err)
		return
	}

	// Create new database user with request provided email and newly hashed password
	newDbUser, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          reqParams.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating user", err)
		return
	}

	// map new database user record to a User struct before marshalling and responding (for JSON tag control purposes - see User struct definition commentary)
	newUser := User{
		ID:          newDbUser.ID,
		CreatedAt:   newDbUser.CreatedAt,
		UpdatedAt:   newDbUser.UpdatedAt,
		Email:       newDbUser.Email,
		IsChirpyRed: newDbUser.IsChirpyRed,
		// Omit the hashed password in response for security purposes
	}

	// respond with HTTP Created (201) success and the new user data
	err = respondWithJSON(w, http.StatusCreated, newUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error occurred", err)
		return
	}
}
