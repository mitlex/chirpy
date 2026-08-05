package main

import "net/http"

// handlerReadinessEndpoint is called by server upon receiving "/healthz" in URL path
// It simply sets and responds with a plain text utf-8 charset Content-Type header and a 200 status code and "OK" body
// It is used to indicate whether or not our server is ready to receive traffic
func handlerReadinessEndpoint(w http.ResponseWriter, r *http.Request) {
	// Header represents the key:value pairs in an HTTP Header map[string][]string
	// We set the Content-Type header using .Set()
	// Nothing is sent to the socket yet, the header is simply buffered
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Flush status 200 and all buffered headers to the socket
	w.WriteHeader(http.StatusOK)

	// Stream the body bytes
	w.Write([]byte("OK"))
}
