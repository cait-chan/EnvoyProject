package main

import (
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// Admin-only backend: reached via Envoy when path starts with /admin or X-Admin: true on /users.

type user struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var users = []user{
	{ID: 1, Name: "Caitlyn", Email: "caitlyn@example.com"},
	{ID: 2, Name: "Johnny", Email: "johnny@example.com"},
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "backend": "admin"})
}

func handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"service": "admin",
		"message": "Admin dashboard (separate backend from public users API)",
		"routes":  []string{"/admin/health", "/admin/dashboard", "/users (via X-Admin header from Envoy)"},
	})
}

func handleUsersAdminView(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"audience": "admin",
		"users":    users,
		"note":     "Same path /users as public API, routed here when client sends X-Admin: true",
	})
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
	log.Fatal("TLS: run scripts/gen-upstream-tls-certs.sh or set TLS_CERT_FILE / TLS_KEY_FILE")
	return "", ""
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /admin/health", handleHealth)
	mux.HandleFunc("GET /admin/dashboard", handleDashboard)
	mux.HandleFunc("GET /users", handleUsersAdminView)

	cert, key := upstreamTLSCertPaths()
	log.Printf("Admin backend on :3010 (HTTPS/TLS) cert=%s key=%s", cert, key)
	srv := &http.Server{
		Addr:    ":3010",
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
