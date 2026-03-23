package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

func upstreamTLSCertPaths() (cert, key string) {
	if c := os.Getenv("TLS_CERT_FILE"); c != "" {
		k := os.Getenv("TLS_KEY_FILE")
		if k == "" {
			log.Fatal("TLS_KEY_FILE must be set when TLS_CERT_FILE is set")
		}
		return c, k
	}
	for _, dir := range []string{filepath.Join("..", "..", "envoy", "certs"), "envoy/certs"} {
		c := filepath.Join(dir, "upstream-server.crt")
		k := filepath.Join(dir, "upstream-server.key")
		if _, err := os.Stat(c); err == nil {
			if _, err2 := os.Stat(k); err2 == nil {
				return c, k
			}
		}
	}
	log.Fatal("TLS: run scripts/gen-upstream-tls-certs.sh or set TLS_CERT_FILE and TLS_KEY_FILE to upstream-server.crt/.key")
	return "", ""
}

func main() {
	// initialize router
	mux := http.NewServeMux()

	// register routes
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /users", handleGetUsers)
	mux.HandleFunc("GET /users/{id}", handleGetUser)

	cert, key := upstreamTLSCertPaths()
	log.Printf("Server starting on :3001 (HTTPS/TLS) cert=%s key=%s", cert, key)
	srv := &http.Server{
		Addr:    ":3001",
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			MaxVersion: tls.VersionTLS13,
		},
	}
	if err := srv.ListenAndServeTLS(cert, key); err != nil {
		log.Fatal(err)
	}
}