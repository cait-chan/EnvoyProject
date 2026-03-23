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

type Product struct {
	ID int `json:"id"`
	Name string `json:"name"`
	Price float64 `json:"price"`
	Stock int `json:"stock"`
}

var products = []Product{
	{ID: 1, Name: "Product 1", Price: 30.0, Stock: 10},
	{ID: 2, Name: "Product 2", Price: 50.0, Stock: 20},
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleGetProducts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func handleGetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	
	for _, product := range products {
		if fmt.Sprintf("%d", product.ID) == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(product)
			return
		}
	}
	http.Error(w, "Product not found", http.StatusNotFound)
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
	mux.HandleFunc("GET /products", handleGetProducts)
	mux.HandleFunc("GET /products/{id}", handleGetProduct)

	cert, key := upstreamTLSCertPaths()
	log.Printf("Server starting on :3003 (HTTPS/TLS) cert=%s key=%s", cert, key)
	srv := &http.Server{
		Addr:    ":3003",
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