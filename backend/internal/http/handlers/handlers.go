package handlers

import (
	"backend/internal/config"
	"backend/internal/database"
	"net/http"
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

func (h *Handlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message":"Service is running"}`))
}

// func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusNotImplemented)
// 	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
// }

// func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusNotImplemented)
// 	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
// }

// Handler implementations are in separate files (auth.go, trips.go)

func (h *Handlers) UpdateTrip(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}

func (h *Handlers) DeleteTrip(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}
