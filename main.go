package main

import (
	"log"
	"net/http"
)

func main() {

	// create new http.ServeMux to route requests
	httpReqRouter := http.NewServeMux()

	// Register handlers for requests
	// WARNING: http.Dir(".") allows the handler to serve any file from the root directory
	// 			fine for this toy project, but better practice to create a dedicated public folder (i.e. http.Dir("public")) to place files we'd like to expose
	// NOTE #1 : FileServer will automatically serve index.html if present in the directory, OR, if not present, serves a listing of all files in the directory
	// NOTE #2 : http.Dir(".") converts the string "." to an http.Dir type.
	// 			 http.Dir implements the http.FileSystem interface, allowing http.FileServer to serve files from the current directory (.).
	httpReqRouter.Handle("/", http.FileServer(http.Dir(".")))

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
