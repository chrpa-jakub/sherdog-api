package event

import (
	"encoding/json"
	"net/http"

	"github.com/chrpa-jakub/sherdog-api/internal/domain/event"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service event.Service
}

func NewRouter(service event.Service) chi.Router {
	handler := Handler{service: service}

	router := chi.NewRouter()
	router.Get("/{id}", handler.get)

	return router
}

func (h Handler) get(w http.ResponseWriter, r *http.Request) {
	event, err := h.service.Get(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "could not fetch event"})
		return
	}

	writeJSON(w, http.StatusOK, event)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
