package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/vkirkizh/travel-map/backend/internal/geocoding"
	"github.com/vkirkizh/travel-map/backend/internal/places"
)

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
	validationErrors := validateCreatePlaceRequest(request)
	if len(validationErrors) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":  "validation failed",
			"fields": validationErrors,
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
