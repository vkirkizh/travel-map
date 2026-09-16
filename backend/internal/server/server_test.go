package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vkirkizh/travel-map/backend/internal/auth"
)

func TestPasswordValidationPreservesWhitespace(t *testing.T) {
	spaces := "      "

	registerErrors := validateRegisterRequest(registerRequest{
		Username:    "traveler",
		Email:       "user@example.com",
		Password:    spaces,
		DisplayName: "Traveler",
	})
	if _, ok := registerErrors["password"]; ok {
		t.Errorf("register rejected a six-character whitespace password: %v", registerErrors)
	}

	loginErrors := validateLoginRequest(loginRequest{
		Email:    "user@example.com",
		Password: " ",
	})
	if _, ok := loginErrors["password"]; ok {
		t.Errorf("login rejected a non-empty whitespace password: %v", loginErrors)
	}

	currentPassword := " current password "
	updateErrors := validateUpdateMeRequest(updateMeRequest{
		DisplayName:     "Traveler",
		Email:           "user@example.com",
		CurrentPassword: &currentPassword,
		NewPassword:     &spaces,
	})
	if _, ok := updateErrors["new_password"]; ok {
		t.Errorf("profile update rejected a six-character whitespace password: %v", updateErrors)
	}

	if got := optionalPassword(&currentPassword); got == nil || *got != currentPassword {
		t.Errorf("optionalPassword() = %v, want exact value %q", got, currentPassword)
	}
}

func TestWriteCurrentUserError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "unauthorized",
			err:        auth.ErrUnauthorized,
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"unauthorized"}`,
		},
		{
			name:       "internal error",
			err:        errors.New("database connection failed"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"internal server error"}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/me", nil)

			writeCurrentUserError(response, request, test.err)

			if response.Code != test.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, test.wantStatus)
			}
			if got := strings.TrimSpace(response.Body.String()); got != test.wantBody {
				t.Errorf("body = %q, want %q", got, test.wantBody)
			}
			if strings.Contains(response.Body.String(), test.err.Error()) && !errors.Is(test.err, auth.ErrUnauthorized) {
				t.Errorf("response exposes internal error: %q", response.Body.String())
			}
		})
	}
}

func TestCurrentUserWithoutCookieIsUnauthorized(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)

	_, err := (&Server{}).currentUser(request)

	if !errors.Is(err, auth.ErrUnauthorized) {
		t.Errorf("currentUser() error = %v, want %v", err, auth.ErrUnauthorized)
	}
}
