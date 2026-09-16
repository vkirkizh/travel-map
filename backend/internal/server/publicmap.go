package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/vkirkizh/travel-map/backend/internal/publicmap"
)

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
