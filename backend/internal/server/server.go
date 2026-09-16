package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
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

type registerRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type createPlaceRequest struct {
	Query string `json:"query"`
}

type updateMeRequest struct {
	DisplayName     string  `json:"display_name"`
	Email           string  `json:"email"`
	CurrentPassword *string `json:"current_password"`
	NewPassword     *string `json:"new_password"`
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

func (s *Server) publicUserMap(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	response, err := s.publicMapRepository.GetByUsername(r.Context(), username)
	if errors.Is(err, publicmap.ErrUserNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{
			"error": "user not found",
		})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "internal server error",
		})
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	request.Username = strings.TrimSpace(request.Username)
	request.Email = auth.NormalizeEmail(request.Email)
	request.DisplayName = strings.TrimSpace(request.DisplayName)

	validationErrors := validateRegisterRequest(request)
	if len(validationErrors) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "validation failed",
			"fields": validationErrors,
		})
		return
	}

	user, sessionToken, err := s.authRepository.Register(
		r.Context(),
		request.Username,
		request.Email,
		request.Password,
		request.DisplayName,
	)
	if errors.Is(err, auth.ErrUserAlreadyExists) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user already exists"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	setSessionCookie(w, sessionToken)

	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	request.Email = auth.NormalizeEmail(request.Email)

	validationErrors := validateLoginRequest(request)
	if len(validationErrors) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "validation failed",
			"fields": validationErrors,
		})
		return
	}

	user, sessionToken, err := s.authRepository.Login(r.Context(), request.Email, request.Password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	setSessionCookie(w, sessionToken)

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("travel_map_session")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		slog.Error(
			"failed to read session cookie",
			"error", err,
			"request_id", middleware.GetReqID(r.Context()),
		)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if err == nil {
		if err := s.authRepository.Logout(r.Context(), cookie.Value); err != nil {
			slog.Error(
				"failed to delete session",
				"error", err,
				"request_id", middleware.GetReqID(r.Context()),
			)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			return
		}
	}

	clearSessionCookie(w)

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) me(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil {
		writeCurrentUserError(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (s *Server) updateMe(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil {
		writeCurrentUserError(w, r, err)
		return
	}

	var request updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}

	request.DisplayName = strings.TrimSpace(request.DisplayName)
	request.Email = auth.NormalizeEmail(request.Email)

	validationErrors := validateUpdateMeRequest(request)
	if len(validationErrors) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "validation failed",
			"fields": validationErrors,
		})
		return
	}

	currentPassword := optionalPassword(request.CurrentPassword)
	newPassword := optionalPassword(request.NewPassword)

	updatedUser, err := s.authRepository.UpdateProfile(r.Context(), auth.UpdateProfileInput{
		UserID:          user.ID,
		DisplayName:     request.DisplayName,
		Email:           request.Email,
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
	})
	if errors.Is(err, auth.ErrCurrentPasswordInvalid) {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "validation failed",
			"fields": map[string]string{
				"current_password": "Current password is incorrect.",
			},
		})
		return
	}
	if errors.Is(err, auth.ErrUserAlreadyExists) {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "user already exists"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": updatedUser})
}

func (s *Server) listPlaces(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil {
		writeCurrentUserError(w, r, err)
		return
	}

	result, err := s.placesRepository.ListByUserID(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"places": result})
}

func (s *Server) createPlace(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil {
		writeCurrentUserError(w, r, err)
		return
	}

	var request createPlaceRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	request.Query = strings.TrimSpace(request.Query)
	if request.Query == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error": "validation failed",
			"fields": map[string]string{
				"query": "Place query is required.",
			},
		})
		return
	}

	resolved, err := s.geocodingService.Resolve(r.Context(), request.Query)
	if errors.Is(err, geocoding.ErrNotFound) {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "place not found",
		})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	created, err := s.placesRepository.Create(r.Context(), user.ID, places.Place{
		Title:       resolved.Title,
		Query:       request.Query,
		CountryCode: resolved.CountryCode,
		Lat:         resolved.Lat,
		Lng:         resolved.Lng,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"place": created})
}

func (s *Server) deletePlace(w http.ResponseWriter, r *http.Request) {
	user, err := s.currentUser(r)
	if err != nil {
		writeCurrentUserError(w, r, err)
		return
	}

	placeID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid place id"})
		return
	}

	if err := s.placesRepository.Delete(r.Context(), user.ID, placeID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) currentUser(r *http.Request) (*auth.User, error) {
	cookie, err := r.Cookie("travel_map_session")
	if errors.Is(err, http.ErrNoCookie) {
		return nil, auth.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	user, err := s.authRepository.CurrentUser(r.Context(), cookie.Value)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func writeCurrentUserError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, auth.ErrUnauthorized) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	slog.Error("failed to get current user", "error", err, "request_id", middleware.GetReqID(r.Context()))
	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func validateRegisterRequest(request registerRequest) map[string]string {
	errs := make(map[string]string)

	username := strings.TrimSpace(request.Username)
	email := auth.NormalizeEmail(request.Email)
	password := request.Password
	displayName := strings.TrimSpace(request.DisplayName)

	if username == "" {
		errs["username"] = "Username is required."
	} else if len(username) < 3 {
		errs["username"] = "Username must be at least 3 characters."
	} else if len(username) > 32 {
		errs["username"] = "Username must be at most 32 characters."
	}

	if email == "" {
		errs["email"] = "Email is required."
	} else if _, err := mail.ParseAddress(email); err != nil {
		errs["email"] = "Email is invalid."
	}

	if password == "" {
		errs["password"] = "Password is required."
	} else if len(password) < 6 {
		errs["password"] = "Password must be at least 6 characters."
	}

	if displayName == "" {
		errs["display_name"] = "Display name is required."
	} else if len(displayName) > 80 {
		errs["display_name"] = "Display name must be at most 80 characters."
	}

	return errs
}

func validateLoginRequest(request loginRequest) map[string]string {
	errs := make(map[string]string)

	email := auth.NormalizeEmail(request.Email)
	password := request.Password

	if email == "" {
		errs["email"] = "Email is required."
	} else if _, err := mail.ParseAddress(email); err != nil {
		errs["email"] = "Email is invalid."
	}

	if password == "" {
		errs["password"] = "Password is required."
	}

	return errs
}

func validateUpdateMeRequest(request updateMeRequest) map[string]string {
	errs := make(map[string]string)

	email := auth.NormalizeEmail(request.Email)

	if request.DisplayName == "" {
		errs["display_name"] = "Display name is required."
	} else if len(request.DisplayName) > 80 {
		errs["display_name"] = "Display name must be at most 80 characters."
	}

	if email == "" {
		errs["email"] = "Email is required."
	} else if _, err := mail.ParseAddress(email); err != nil {
		errs["email"] = "Email is invalid."
	}

	newPassword := optionalPassword(request.NewPassword)
	currentPassword := optionalPassword(request.CurrentPassword)

	if newPassword != nil {
		if len(*newPassword) < 6 {
			errs["new_password"] = "New password must be at least 6 characters."
		}

		if currentPassword == nil {
			errs["current_password"] = "Current password is required to change password."
		}
	}

	return errs
}

func optionalPassword(value *string) *string {
	if value == nil || *value == "" {
		return nil
	}
	return value
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "travel_map_session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "travel_map_session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
