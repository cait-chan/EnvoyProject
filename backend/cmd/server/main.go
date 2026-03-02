package main

import (
	"log"
	"net/http"
	"github.com/cait-chan/EnvoyProject/backend/internal/api"
	"github.com/cait-chan/EnvoyProject/backend/internal/envoy"
)

func main() {
	adminURL := "http://localhost:9901"
	envoyClient := envoy.NewClient(adminURL)
	handler := api.NewHandler(envoyClient)
	router := api.NewRouter(handler)

	log.Println("Backend server starting on :8080")
	err := http.ListenAndServe(":8080", router)
	if err != nil {
		log.Fatal(err)
	}
}