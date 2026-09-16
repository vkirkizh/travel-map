package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/vkirkizh/travel-map/backend/internal/auth"
)

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
