package main

import (
	"backend/internal/config"
	"backend/internal/database"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"backend/internal/http/handlers"
	"backend/internal/services"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	_ "backend/docs"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// @title           Itinera API
// @version         1.0
// @description     API to plan trip and manage expenses Itinera.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found, using system environment variables")
	}

	cfg := config.Load()
	cfg.LogSummary()

	pool, err := database.NewPostgress(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	tripsRepo := database.NewTripRepository(pool)
	placesRepo := database.NewPlaceRepository(pool)
	expensesRepo := database.NewExpenseRepository(pool)
	authRepo := database.NewAuthRepository(pool)
	activityRepo := database.NewActivityRepository(pool)

	exchangeRateSvc := services.NewExchangeRateService(pool)
	expenseSvc := services.NewExpenseService(tripsRepo, expensesRepo, exchangeRateSvc)

	h := handlers.NewHandlers(tripsRepo, placesRepo, expensesRepo, authRepo, activityRepo, expenseSvc, cfg)

	router := setupRouter(cfg, h)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}
	log.Println("✅ Server exited gracefully")

}

func setupRouter(cfg *config.Config, h *handlers.Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowOriginFunc: func(r *http.Request, origin string) bool {
			// Permite orígenes de desarrollo local
			if origin == "http://localhost:5173" || origin == "http://127.0.0.1:5173" || origin == "http://localhost:3000" {
				return true
			}
			// Producción Vercel (todos los subdominios)
			if strings.Contains(origin, "vercel.app") {
				return true
			}
			// Permite cualquier subdominio de vercel.app (preview deployments)
			if strings.HasSuffix(origin, ".vercel.app") {
				return true
			}

			if strings.Contains(origin, "ngrok-free.app") || strings.Contains(origin, "ngrok.io") {
            return true
        }
			return false
		},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Origin", "X-Requested-With", "ngrok-skip-browser-warning"},
		ExposedHeaders:   []string{"Link", "Content-Length"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.TripsRepo.Pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"unhealthy","database":"disconnected"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"healthy","database":"connected"}`))
	})

	r.Route("/api/v1", func(router chi.Router) {
		handlers.RegisterApiRoutes(router, h)
	})

	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), // The url pointing to API definition
	))

	return r
}
