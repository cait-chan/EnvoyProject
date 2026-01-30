package main

import (
	"encoding/json"
	"log"
	"net/http"
	"fmt"
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

func main() {
	// initialize router
	mux := http.NewServeMux()

	// register routes
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /products", handleGetProducts)
	mux.HandleFunc("GET /products/{id}", handleGetProduct)

	// start server
	log.Println("Server starting on :3003")
	err := http.ListenAndServe(":3003", mux)
	if err != nil {
		log.Fatal(err)
	}
}