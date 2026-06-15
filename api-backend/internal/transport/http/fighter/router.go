package fighter

import (
	"encoding/json"
	"net/http"

	"github.com/chrpa-jakub/sherdog-api/internal/domain/fighter"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service fighter.Service
}

func NewRouter(service fighter.Service) chi.Router {
	handler := Handler{service: service}

	router := chi.NewRouter()
	router.Get("/{id}", handler.get)

	return router
}

func (h Handler) get(w http.ResponseWriter, r *http.Request) {
	fighter, err := h.service.Get(r.Context(), fighter.ShortenID(chi.URLParam(r, "id")))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "could not fetch fighter"})
		return
	}

	writeJSON(w, http.StatusOK, fighter)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
