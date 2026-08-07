package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// respondWithJson marshals and returns a payload as JSON
// args: calling handler's ResponseWriter, an HTTP status code, and a payload to be marshalled (typically a struct)
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) error {
	response, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
	return nil
}

// respondWithError wraps respondWithJson to always send an error message
// use this when something goes wrong and you need to respond with an error
// args: calling handler's ResponseWriter, HTTP status code, error message, underlying error
func respondWithError(w http.ResponseWriter, code int, msg string, err error) error {
	log.Printf("Responding with error: %v, underlying error: %v", msg, err) // for internal server logging; log.Printf also provides a timestamp for the error when printed to the console
	return respondWithJSON(w, code, map[string]string{"error": msg})        // the client only receives this error message, not the one from the line above
}
