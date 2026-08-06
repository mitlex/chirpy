package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

// apiConfig will hold any stateful in-memory data that needs tracking
type apiConfig struct {
	fileserverHits atomic.Int32 // tracks how many requests are made to our website - this type allows safe incrementing and reading of an integer value across multiple goroutines (HTTP requests)
}

// middlewareMetricsInc increments the fileserverHits counter every time middlewareMetricsInc is called
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// handlerDisplaySiteHits writes the number of requests that have been counted as plain text to the HTTP response
func (cfg *apiConfig) handlerDisplaySiteHitsEndpoint(w http.ResponseWriter, r *http.Request) {
	hits := fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load()) // .Load() safely reads the fileserverHits counter current value
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(hits))
}

// handlerResetSiteHits resets the fileserverHits back to 0
func (cfg *apiConfig) handlerResetSiteHitsEndpoint(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Swap(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func main() {

	// Create new http.ServeMux to route requests
	httpReqRouter := http.NewServeMux()

	// Instantiate apiConfig for stateful data tracking
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	// Register handlers for requests
	// FileServer (/app/)
	// WARNING: http.Dir(".") allows the handler to serve any file from the process's current working directory (.).
	// 			Fine for this toy project, but better practice to create a dedicated public folder (i.e. http.Dir("public")) to place files we'd like to expose.
	//
	// NOTE #1: FileServer will automatically serve index.html if present in the directory, OR, if not present, serves a listing of all files in the directory.
	//
	// NOTE #2: http.Dir(filepathRoot) converts the string "." to an http.Dir type.
	// 			http.Dir implements the http.FileSystem interface, allowing http.FileServer to serve files from the current working directory.
	//			In other words, http.Dir tells the FileServer handler where on disk to look for files.
	//
	// NOTE #3: StripPrefix strips /app from the URL path before FileServer looks for it on disk.
	//
	// 			Example: URL request /app/assets/logo.png becomes /assets/logo.png which is where the logo.png file actually lives on disk (reminder: we don't have an /app folder on disk).
	//
	// 			This allows us to create an /app namespace that the httpReqRouter Mux can route /app/ URL traffic to, even though the FileServer is serving files from current working directory (e.g. root directory).
	// 			If we continued to register the FileServer at "/" then any requests like /healthz would be routed to the FileServer as well, as "/" swallows all other requests like "/x", "/x/y".
	//
	// NOTE #4: We register "/app/" as the pattern so httpReqRouter routes the entire subtree
	//         	of URLs (e.g. /app/index.html, /app/assets/logo.png) to the FileServer handler.
	//          If we registered "/app" (no trailing slash), httpReqRouter would only match
	//          the exact literal path "/app" — nothing nested underneath it would reach the FileServer handler.
	//			Go can redirect /app to /app/ when "/app/" is registered.
	const filepathRoot = "."
	httpReqRouter.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))) // wrap FileServer to track # site requests

	// Readiness endpoint (/healthz)
	// Notice how we register this pattern using HandleFunc instead of Handle - look at the parameter types that Handle accepts vs. HandleFunc
	httpReqRouter.HandleFunc("GET /healthz", handlerReadinessEndpoint)              // /healthz only accepts GET requests, server should return 405 (method not allowed) response automatically if other method used
	httpReqRouter.HandleFunc("GET /metrics", apiCfg.handlerDisplaySiteHitsEndpoint) // /metrics only accepts GET requests
	httpReqRouter.HandleFunc("POST /reset", apiCfg.handlerResetSiteHitsEndpoint)    // only accepts POST requests

	// create http server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: httpReqRouter, // server passes requests to this master Handler, which then routes the requests based on the URL
	}

	// start the server, listening on Addr (:8080)
	// only returns error when it stops or fails to start
	// it blocks the goroutine (main) it's running in for the life of the server hence we place it at the end of our main func
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal(err) // this also calls os.Exit(1) terminating the program immediately; if server can't run, nothing left for main to do
	}
}
