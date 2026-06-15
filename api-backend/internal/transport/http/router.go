package http

import (
	"github.com/chrpa-jakub/sherdog-api/internal/domain/event"
	"github.com/chrpa-jakub/sherdog-api/internal/domain/fighter"
	eventhttp "github.com/chrpa-jakub/sherdog-api/internal/transport/http/event"
	fighterhttp "github.com/chrpa-jakub/sherdog-api/internal/transport/http/fighter"
	"github.com/go-chi/chi/v5"
)

func NewRouter(fighterService fighter.Service, eventService event.Service) *chi.Mux {
	router := chi.NewRouter()
	router.Mount("/fighter", fighterhttp.NewRouter(fighterService))
	router.Mount("/events", eventhttp.NewRouter(eventService))

	return router
}
