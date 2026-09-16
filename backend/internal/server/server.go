package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vkirkizh/travel-map/backend/internal/auth"
	"github.com/vkirkizh/travel-map/backend/internal/config"
	"github.com/vkirkizh/travel-map/backend/internal/geocoding"
	"github.com/vkirkizh/travel-map/backend/internal/places"
	"github.com/vkirkizh/travel-map/backend/internal/publicmap"
)

type Server struct {
	db                  *pgxpool.Pool
	publicMapRepository *publicmap.Repository
	authRepository      *auth.Repository
	placesRepository    *places.Repository
	geocodingService    *geocoding.Service
}

func New(db *pgxpool.Pool, cfg config.Config) http.Handler {
	s := &Server{
		db:                  db,
		publicMapRepository: publicmap.NewRepository(db),
		authRepository:      auth.NewRepository(db),
		placesRepository:    places.NewRepository(db),
		geocodingService:    geocoding.NewService(db),
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:3000",
		},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/healthz", s.healthz)
	r.Get("/readyz", s.readyz)

	r.Route("/api", func(r chi.Router) {
		r.Get("/public/users/{username}/map", s.publicUserMap)

		r.Post("/auth/register", s.register)
		r.Post("/auth/login", s.login)
		r.Post("/auth/logout", s.logout)

		r.Get("/me", s.me)
		r.Patch("/me", s.updateMe)

		r.Get("/places", s.listPlaces)
		r.Post("/places", s.createPlace)
		r.Delete("/places/{id}", s.deletePlace)
	})

	return r
}

func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) readyz(w http.ResponseWriter, r *http.Request) {
	if err := s.db.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"error":  "database is not available",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeCurrentUserError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, auth.ErrUnauthorized) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	slog.Error("failed to get current user", "error", err, "request_id", middleware.GetReqID(r.Context()))
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func optionalPassword(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	return value
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
