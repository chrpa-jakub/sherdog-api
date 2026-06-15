package http

import (
	"log"
	"net/http"
	"time"

	"github.com/chrpa-jakub/sherdog-api/internal/domain/event"
	"github.com/chrpa-jakub/sherdog-api/internal/domain/fighter"
	eventhttp "github.com/chrpa-jakub/sherdog-api/internal/transport/http/event"
	fighterhttp "github.com/chrpa-jakub/sherdog-api/internal/transport/http/fighter"
	"github.com/go-chi/chi/v5"
)

func NewRouter(fighterService fighter.Service, eventService event.Service) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logRequests)
	router.Mount("/fighter", fighterhttp.NewRouter(fighterService))
	router.Mount("/events", eventhttp.NewRouter(eventService))

	return router
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(payload []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	return r.ResponseWriter.Write(payload)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{ResponseWriter: w}

		next.ServeHTTP(recorder, r)

		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		log.Printf(
			"request completed method=%s path=%s status=%d duration=%s remote_addr=%s",
			r.Method,
			r.URL.Path,
			status,
			time.Since(start),
			r.RemoteAddr,
		)
	})
}
