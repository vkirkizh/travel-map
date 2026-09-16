package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/vkirkizh/travel-map/backend/internal/auth"
)

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
