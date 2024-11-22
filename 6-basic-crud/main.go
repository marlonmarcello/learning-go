package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	api := &api{addr: ":8080"}

	mux := http.NewServeMux()

	srv := &http.Server{
		Addr:    api.addr,
		Handler: mux,
	}

	// catch all
	mux.HandleFunc("/", api.handleRoot)

	// post users
	mux.HandleFunc("POST /users", api.handlePostUsers)

	// get users
	mux.HandleFunc("GET /users/{id}", api.handleGetUsers)

	// delete users
	mux.HandleFunc("DELETE /users/{id}", api.handleDeleteUsers)

	// start
	fmt.Println("Server listening on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
