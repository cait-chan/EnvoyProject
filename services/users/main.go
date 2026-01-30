package main

import (
	"encoding/json"
	"log"
	"net/http"
	"fmt"
)

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Email string `json:"email"`
}

var users = []User{
	{ID: 1, Name: "Caitlyn", Email: "caitlyn@example.com"},
	{ID: 2, Name: "Johnny", Email: "johnny@example.com"},
}


func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGetUsers(w http.ResponseWriter, r *http.Request) {
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	for _, user := range users {
		if fmt.Sprintf("%d", user.ID) == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(user)
			return
		}
	}

	http.Error(w, "User not found", http.StatusNotFound)
}

func main() {
	// initialize router
	mux := http.NewServeMux()

	// register routes
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /users", handleGetUsers)
	mux.HandleFunc("GET /users/{id}", handleGetUser)

	// start server
	log.Println("Server starting on :3001")
	err := http.ListenAndServe(":3001", mux)
	if err != nil {
		log.Fatal(err)
	}
}