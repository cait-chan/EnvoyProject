package main

import (
	"encoding/json"
	"log"
	"net/http"
	"fmt"
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

func main() {
	// initialize router
	mux := http.NewServeMux()

	// register routes
	mux.HandleFunc("GET /health", handleHealth)
	mux.HandleFunc("GET /orders", handleGetOrders)
	mux.HandleFunc("GET /orders/{id}", handleGetOrder)

	// start server
	log.Println("Server starting on :3002")
	err := http.ListenAndServe(":3002", mux)
	if err != nil {
		log.Fatal(err)
	}
}