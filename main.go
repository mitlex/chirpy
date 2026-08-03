package main

import (
	"log"
	"net/http"
)

func main() {

	// create new http.ServeMux to route requests
	httpReqRouter := http.NewServeMux()

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
