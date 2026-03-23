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

type Order struct {
	ID int `json:"id"`
	UserID int `json:"user_id"`
	Total float64 `json:"total"`
	Status string `json:"status"`
}

var orders = []Order{
	{ID: 1, UserID: 1, Total: 100.0, Status: "pending"},
	{ID: 2, UserID: 2, Total: 200.0, Status: "completed"},
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGetOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func handleGetOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	for _, order := range orders {
		if fmt.Sprintf("%d", order.ID) == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(order)
			return
		}
	}

	http.Error(w, "Order not found", http.StatusNotFound)
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
	mux.HandleFunc("GET /orders", handleGetOrders)
	mux.HandleFunc("GET /orders/{id}", handleGetOrder)

	cert, key := upstreamTLSCertPaths()
	log.Printf("Server starting on :3002 (HTTPS/TLS) cert=%s key=%s", cert, key)
	srv := &http.Server{
		Addr:    ":3002",
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