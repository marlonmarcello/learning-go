package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
)

type User struct {
	Name string `json:"name"`
}

var userStore = make(map[int]User)

var storeMutex sync.RWMutex

func main() {
	mux := http.NewServeMux()

	// catch all
	mux.HandleFunc("/", handleRoot)

	// post users
	mux.HandleFunc("POST /users", handlePostUsers)

	// get users
	mux.HandleFunc("GET /users/{id}", handleGetUsers)

	// delete users
	mux.HandleFunc("DELETE /users/{id}", handleDeleteUsers)

	// start
	fmt.Println("Server listening on port 8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world\n")
}

func handlePostUsers(w http.ResponseWriter, r *http.Request) {
	var user User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	storeMutex.Lock()
	userStore[len(userStore)+1] = user
	storeMutex.Unlock()

	jsonUser, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-type", "application/json")
	w.Write(jsonUser)
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	storeMutex.RLock()
	user, ok := userStore[id]
	storeMutex.RUnlock()

	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	jsonUser, err := json.Marshal(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonUser)
}

func handleDeleteUsers(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, ok := userStore[id]; !ok {
		http.Error(w, "user not found", http.StatusInternalServerError)
		return
	}

	storeMutex.Lock()
	delete(userStore, id)
	storeMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}
