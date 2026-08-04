package main

import (
	"log"
	"net/http"
)

func main() {

	// create new http.ServeMux to route requests
	httpReqRouter := http.NewServeMux()

	// Register handlers for requests

	// FileServer
	// WARNING: http.Dir(".") allows the handler to serve any file from the root directory
	// 			Fine for this toy project, but better practice to create a dedicated public folder (i.e. http.Dir("public")) to place files we'd like to expose
	// NOTE #1: FileServer will automatically serve index.html if present in the directory, OR, if not present, serves a listing of all files in the directory
	// NOTE #2: http.Dir(filepathRoot) converts the string "." to an http.Dir type
	// 			http.Dir implements the http.FileSystem interface, allowing http.FileServer to serve files from the current directory (root: .)
	//			In other words, http.Dir tells the FileServer where on disk to look for files
	// NOTE #3: StripPrefix strips `/app` from the URL path before FileServer looks for it on disk
	// 			Example: URL request /app/assets/logo.png becomes /assets/logo.png which is where the logo.png file actually lives on disk (reminder: we don't have an /app folder on disk)
	// 			This allows us to create a /app namespace that the httpReqRouter Mux can route /app URL traffic to, even though the FileServer is serving files from root (.) directory.
	// 			If we continued to register the FileServer at "/" then any requests like /healthz would be routed to the FileServer as well, as "/" swallows all other requests like "/x", "/x/y"
	const filepathRoot = "."
	httpReqRouter.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))

	// Readiness endpoint (/healthz)
	httpReqRouter.HandleFunc("/healthz", serverReadinessEndpoint)

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
