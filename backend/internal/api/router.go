package api

import (
	"net/http"
	"github.com/cait-chan/EnvoyProject/backend/internal/envoy"
)

type Handler struct {
	envoyClient *envoy.Client
}

func NewHandler(envoyClient *envoy.Client) *Handler {
	return &Handler{
		envoyClient: envoyClient,
	}
}

func NewRouter(handler *Handler) *http.ServeMux {
	// create router
	mux := http.NewServeMux()

	// register routes
	mux.HandleFunc("GET /api/stats", handler.handleGetStats)
	mux.HandleFunc("GET /api/clusters", handler.handleGetClusters)

	return mux
}

func (h *Handler) handleGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.envoyClient.GetStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(stats)
}

func (h *Handler) handleGetClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := h.envoyClient.GetClusters()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(clusters)
}