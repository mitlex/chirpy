package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq" // we must import this postgres driver even though we don't use it in our code - it's just being imported for "side effects" such that our program knows how to talk to our database
	"github.com/mitlex/chirpy/internal/database"
)

// apiConfig will hold any stateful in-memory data that needs tracking
type apiConfig struct {
	fileserverHits atomic.Int32      // tracks how many requests are made to our website - this type allows safe incrementing and reading of an integer value across multiple goroutines (HTTP requests)
	db             *database.Queries // to give handlers access to database queries
	platform       string            // to represent environment name for purposes such as restricting handlers to execute only in certain environments
	jwtSecret      string            // used for signing and validating JSON Web Tokens
	polkaKey       string            // Polka API Key - Polka will send this with each Chirpy Red subscription webhook to prove the webhook is coming from Polka and not someone trying to get a free Chirpy Red subscription
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
	hits := fmt.Sprintf(
		`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`,
		cfg.fileserverHits.Load()) // .Load() safely reads the fileserverHits counter current value
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(hits))
}

// handlerResetEndpoint resets the fileserverHits back to 0 and deletes all users from the database
func (cfg *apiConfig) handlerResetEndpoint(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "forbidden request", nil)
		return
	}
	cfg.fileserverHits.Swap(0)
	err := cfg.db.ResetUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error resetting users", err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func main() {
	godotenv.Load() // load .env file into environment variables

	dbUrl := os.Getenv("DB_URL") // get database URL from environment

	// open connection to database
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db) // use SQLC generated database package to create a new *database.Queries

	httpReqRouter := http.NewServeMux() // Create new http.ServeMux to route requests

	// Instantiate apiConfig for stateful data tracking
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       os.Getenv("PLATFORM"),
		jwtSecret:      os.Getenv("JWT_SECRET"),
		polkaKey:       os.Getenv("POLKA_KEY"),
	}

	// Register handlers for requests
	// FileServer endpoint (/app/)
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
	// 			If we continued to register the FileServer at "/" then any unregisted request paths like /xyz would be routed to the FileServer as well,
	// 			be processed by the FileServer logic, which would look for a file xyz, would not find it, and then return a 404 error.
	//
	// NOTE #4: We register "/app/" as the pattern so httpReqRouter routes the entire subtree
	//         	of URLs (e.g. /app/index.html, /app/assets/logo.png) to the FileServer handler.
	//          If we registered "/app" (no trailing slash), httpReqRouter would only match
	//          the exact literal path "/app" — nothing nested underneath it would reach the FileServer handler.
	//			Go can redirect /app to /app/ when "/app/" is registered.
	const filepathRoot = "."
	httpReqRouter.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))) // wrap FileServer to track # site requests

	// API endpoints
	// Notice how we register this pattern using HandleFunc instead of Handle - look at the parameter types that Handle accepts vs. HandleFunc
	httpReqRouter.HandleFunc("GET /api/healthz", handlerReadinessEndpoint)                // only accepts GET requests, server should return 405 (method not allowed) response automatically if other method used
	httpReqRouter.HandleFunc("GET /admin/metrics", apiCfg.handlerDisplaySiteHitsEndpoint) // only accepts GET requests
	httpReqRouter.HandleFunc("POST /admin/reset", apiCfg.handlerResetEndpoint)            // only accepts POST requests
	httpReqRouter.HandleFunc("POST /api/chirps", apiCfg.handlerCreateChirp)
	httpReqRouter.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)
	httpReqRouter.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)
	httpReqRouter.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	httpReqRouter.HandleFunc("POST /api/login", apiCfg.handlerLoginUser)
	httpReqRouter.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)
	httpReqRouter.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)
	httpReqRouter.HandleFunc("PUT /api/users", apiCfg.handlerUpdateEmailAndPassword)
	httpReqRouter.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)
	httpReqRouter.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerUpgradeUserToChirpyRed)

	// create http server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: httpReqRouter, // server passes requests to this master Handler, which then routes the requests based on the URL
	}

	// start the server, listening on Addr (:8080)
	// only returns error when it stops or fails to start
	// it blocks the goroutine (main) it's running in for the life of the server hence we place it at the end of our main func
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err) // this also calls os.Exit(1) terminating the program immediately; if server can't run, nothing left for main to do
	}
}
