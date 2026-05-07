package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/http/middleware"
	"backend/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTripContextStore struct {
	getForkFunc     func(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
	getTripMetaFunc func(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error)
	forkTripFunc     func(ctx context.Context, originalTripID string, userID *uuid.UUID, sessionID *string) (*models.Trip, error)
}

func (m *mockTripContextStore) GetFork(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
	if m.getForkFunc != nil {
		return m.getForkFunc(ctx, forkedFrom, userID, sessionID)
	}
	return nil, pgx.ErrNoRows
}

func (m *mockTripContextStore) GetTripMeta(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error) {
	if m.getTripMetaFunc != nil {
		return m.getTripMetaFunc(ctx, tripID, userID, sessionID)
	}
	return false, false, nil
}

func (m *mockTripContextStore) ForkTrip(ctx context.Context, originalTripID string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
	if m.forkTripFunc != nil {
		return m.forkTripFunc(ctx, originalTripID, userID, sessionID)
	}
	return nil, nil
}

func TestResolveTripContext_Owner(t *testing.T) {
	sessionID := "test-session-owner"
	demoTripID := uuid.New().String()

	store := &mockTripContextStore{
		getTripMetaFunc: func(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error) {
			return true, false, nil
		},
	}

	var capturedReq *http.Request
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	})

	r := chi.NewRouter()
	r.With(middleware.ResolveTripContext(store)).Group(func(r chi.Router) {
		r.Get("/trips/{id}", func(w http.ResponseWriter, r *http.Request) {
			nextHandler.ServeHTTP(w, r)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/trips/"+demoTripID, nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.ContextKeySessionId{}, sessionID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "Should return 200")
	require.NotNil(t, capturedReq, "Next handler should be called")

	gotTripID := middleware.GetWorkingTripID(capturedReq)
	assert.Equal(t, demoTripID, gotTripID, "Should use original trip ID for owner")
}

func TestResolveTripContext_DemoGuestNoFork(t *testing.T) {
	sessionID := "test-session-guest"
	demoTripID := uuid.New().String()
	forkTripID := uuid.New().String()

	store := &mockTripContextStore{
		getForkFunc: func(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
			return nil, pgx.ErrNoRows
		},
		getTripMetaFunc: func(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error) {
			return false, true, nil
		},
		forkTripFunc: func(ctx context.Context, originalTripID string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
			newTrip := &models.Trip{ID: uuid.MustParse(forkTripID)}
			return newTrip, nil
		},
	}

	var capturedReq *http.Request
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	})

	r := chi.NewRouter()
	r.With(middleware.ResolveTripContext(store)).Group(func(r chi.Router) {
		r.Post("/trips/{id}/*", func(w http.ResponseWriter, r *http.Request) {
			nextHandler.ServeHTTP(w, r)
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/trips/"+demoTripID+"/places", nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.ContextKeySessionId{}, sessionID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "Should return 200")
	require.NotNil(t, capturedReq, "Next handler should be called")

	gotTripID := middleware.GetWorkingTripID(capturedReq)
	assert.Equal(t, forkTripID, gotTripID, "Should use forked trip ID")
}

func TestResolveTripContext_DemoGuestWithFork(t *testing.T) {
	sessionID := "test-session-returning"
	demoTripID := uuid.New().String()
	existingForkID := uuid.New().String()

	store := &mockTripContextStore{
		getForkFunc: func(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
			return &models.Trip{ID: uuid.MustParse(existingForkID)}, nil
		},
	}

	var capturedReq *http.Request
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	})

	r := chi.NewRouter()
	r.With(middleware.ResolveTripContext(store)).Group(func(r chi.Router) {
		r.Post("/trips/{id}/*", func(w http.ResponseWriter, r *http.Request) {
			nextHandler.ServeHTTP(w, r)
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/trips/"+demoTripID+"/activities", nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.ContextKeySessionId{}, sessionID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "Should return 200")
	require.NotNil(t, capturedReq, "Next handler should be called")

	gotTripID := middleware.GetWorkingTripID(capturedReq)
	assert.Equal(t, existingForkID, gotTripID, "Should reuse existing fork")
}

func TestResolveTripContext_DemoUserNoFork(t *testing.T) {
	userID := uuid.New()
	demoTripID := uuid.New().String()
	forkTripID := uuid.New().String()

	store := &mockTripContextStore{
		getForkFunc: func(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
			return nil, pgx.ErrNoRows
		},
		getTripMetaFunc: func(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error) {
			return false, true, nil
		},
		forkTripFunc: func(ctx context.Context, originalTripID string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
			newTrip := &models.Trip{ID: uuid.MustParse(forkTripID)}
			return newTrip, nil
		},
	}

	var capturedReq *http.Request
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	})

	r := chi.NewRouter()
	r.With(middleware.ResolveTripContext(store)).Group(func(r chi.Router) {
		r.Put("/trips/{id}", func(w http.ResponseWriter, r *http.Request) {
			nextHandler.ServeHTTP(w, r)
		})
	})

	req := httptest.NewRequest(http.MethodPut, "/trips/"+demoTripID, nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.ContextKeyUserId{}, userID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "Should return 200")
	require.NotNil(t, capturedReq, "Next handler should be called")

	gotTripID := middleware.GetWorkingTripID(capturedReq)
	assert.Equal(t, forkTripID, gotTripID, "Should create fork with user_id")
}

func TestResolveTripContext_PrivateTrip(t *testing.T) {
	userID := uuid.New()
	privateTripID := uuid.New().String()

	store := &mockTripContextStore{
		getForkFunc: func(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
			return nil, pgx.ErrNoRows
		},
		getTripMetaFunc: func(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error) {
			return false, false, nil
		},
	}

	var capturedReq *http.Request
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	})

	r := chi.NewRouter()
	r.With(middleware.ResolveTripContext(store)).Group(func(r chi.Router) {
		r.Get("/trips/{id}", func(w http.ResponseWriter, r *http.Request) {
			nextHandler.ServeHTTP(w, r)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/trips/"+privateTripID, nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.ContextKeyUserId{}, userID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "Should return 200 (403 handled by handler)")
	require.NotNil(t, capturedReq, "Next handler should be called")

	gotTripID := middleware.GetWorkingTripID(capturedReq)
	assert.Equal(t, privateTripID, gotTripID, "Should pass through for private trip")
}

func TestResolveTripContext_NotFound(t *testing.T) {
	sessionID := "test-session"
	tripID := uuid.New().String()

	store := &mockTripContextStore{
		getForkFunc: func(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
			return nil, pgx.ErrNoRows
		},
		getTripMetaFunc: func(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error) {
			return false, false, pgx.ErrNoRows
		},
	}

	var capturedReq *http.Request
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	})

	r := chi.NewRouter()
	r.With(middleware.ResolveTripContext(store)).Group(func(r chi.Router) {
		r.Get("/trips/{id}", func(w http.ResponseWriter, r *http.Request) {
			nextHandler.ServeHTTP(w, r)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/trips/"+tripID, nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.ContextKeySessionId{}, sessionID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code, "Should return 404")
	require.Nil(t, capturedReq, "Next handler should NOT be called")
}

func TestResolveTripContext_GetOnDemo(t *testing.T) {
	sessionID := "test-session-read"
	demoTripID := uuid.New().String()

	store := &mockTripContextStore{
		getForkFunc: func(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
			return nil, pgx.ErrNoRows
		},
		getTripMetaFunc: func(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error) {
			return false, true, nil
		},
	}

	var capturedReq *http.Request
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedReq = r
	})

	r := chi.NewRouter()
	r.With(middleware.ResolveTripContext(store)).Group(func(r chi.Router) {
		r.Get("/trips/{id}", func(w http.ResponseWriter, r *http.Request) {
			nextHandler.ServeHTTP(w, r)
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/trips/"+demoTripID, nil)
	req = req.WithContext(context.WithValue(context.Background(), middleware.ContextKeySessionId{}, sessionID))
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code, "Should return 200")
	require.NotNil(t, capturedReq, "Next handler should be called")

	gotTripID := middleware.GetWorkingTripID(capturedReq)
	assert.Equal(t, demoTripID, gotTripID, "Should use original demo ID for GET")
}