package http

import (
	"backend/internal/config"
	"backend/internal/database"
	"net/http"

	"github.com/go-chi/chi"
)

type Handlers struct {
	DB     *database.DB
	Config *config.Config
}

func NewHandlers(db *database.DB, cfg *config.Config) *Handlers {
	return &Handlers{
		DB:     db,
		Config: cfg,
	}
}

func ApiRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Post("/auth/register", h.Register)
	r.Post("/auth/login", h.Login)

	r.Group(func(r chi.Router) {
		r.Get("/trips", h.ListTrips)
		r.Post("/trips", h.CreateTrip)
		r.Get("/trips/{id}", h.GetTrip)
		r.Put("/trips/{id}", h.UpdateTrip)
		r.Delete("/trips/{id}", h.DeleteTrip)
	})

	return r
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}

func (h *Handlers) ListTrips(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}

func (h *Handlers) CreateTrip(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}

func (h *Handlers) GetTrip(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}

func (h *Handlers) UpdateTrip(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}

func (h *Handlers) DeleteTrip(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}
