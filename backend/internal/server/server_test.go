package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vkirkizh/travel-map/backend/internal/auth"
)

func TestPlaceQueryLengthValidation(t *testing.T) {
	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{name: "below minimum", query: "ab", wantErr: true},
		{name: "exact minimum", query: "abc", wantErr: false},
		{name: "exact maximum", query: strings.Repeat("a", 200), wantErr: false},
		{name: "above maximum", query: strings.Repeat("a", 201), wantErr: true},
		{name: "unicode counts as characters", query: strings.Repeat("界", 200), wantErr: false},
		{name: "trims before counting", query: " \tabc\n ", wantErr: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := validateCreatePlaceRequest(createPlaceRequest{Query: test.query})
			_, gotErr := errs["query"]
			if gotErr != test.wantErr {
				t.Errorf("query error present = %t, want %t; errors: %v", gotErr, test.wantErr, errs)
			}
		})
	}
}

func TestDisplayNameLengthValidation(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		wantErr     bool
	}{
		{name: "below minimum", displayName: "a", wantErr: true},
		{name: "exact minimum", displayName: "ab", wantErr: false},
		{name: "exact maximum", displayName: strings.Repeat("a", 50), wantErr: false},
		{name: "above maximum", displayName: strings.Repeat("a", 51), wantErr: true},
		{name: "unicode counts as characters", displayName: strings.Repeat("界", 50), wantErr: false},
		{name: "trims before counting", displayName: " \tab\n ", wantErr: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registerErrors := validateRegisterRequest(registerRequest{
				Username:    "traveler",
				Email:       "user@example.com",
				Password:    "secret",
				DisplayName: test.displayName,
			})
			_, registerErr := registerErrors["display_name"]

			updateErrors := validateUpdateMeRequest(updateMeRequest{
				DisplayName: test.displayName,
				Email:       "user@example.com",
			})
			_, updateErr := updateErrors["display_name"]

			if registerErr != test.wantErr || updateErr != test.wantErr {
				t.Errorf(
					"display name errors: register = %t, update = %t, want %t",
					registerErr,
					updateErr,
					test.wantErr,
				)
			}
		})
	}
}

func TestNewPasswordLengthValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "below minimum", password: "12345", wantErr: true},
		{name: "exact minimum", password: "123456", wantErr: false},
		{name: "exact maximum", password: strings.Repeat("a", 64), wantErr: false},
		{name: "above maximum", password: strings.Repeat("a", 65), wantErr: true},
		{name: "six bytes across three unicode characters", password: "ééé", wantErr: false},
		{name: "whitespace is preserved", password: "      ", wantErr: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registerErrors := validateRegisterRequest(registerRequest{
				Username:    "traveler",
				Email:       "user@example.com",
				Password:    test.password,
				DisplayName: "Traveler",
			})
			_, registerErr := registerErrors["password"]

			currentPassword := " current password "
			newPassword := test.password
			updateErrors := validateUpdateMeRequest(updateMeRequest{
				DisplayName:     "Traveler",
				Email:           "user@example.com",
				CurrentPassword: &currentPassword,
				NewPassword:     &newPassword,
			})
			_, updateErr := updateErrors["new_password"]

			if registerErr != test.wantErr || updateErr != test.wantErr {
				t.Errorf(
					"password errors: register = %t, update = %t, want %t",
					registerErr,
					updateErr,
					test.wantErr,
				)
			}
		})
	}
}

func TestLoginPasswordValidationRemainsNonEmptyOnly(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{name: "empty", password: "", wantErr: true},
		{name: "one byte", password: "x", wantErr: false},
		{name: "whitespace", password: " ", wantErr: false},
		{name: "above creation maximum", password: strings.Repeat("a", 65), wantErr: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := validateLoginRequest(loginRequest{
				Email:    "user@example.com",
				Password: test.password,
			})
			_, gotErr := errs["password"]
			if gotErr != test.wantErr {
				t.Errorf("password error present = %t, want %t; errors: %v", gotErr, test.wantErr, errs)
			}
		})
	}
}

func TestOptionalPasswordPreservesExactValue(t *testing.T) {
	password := " current password "

	if got := optionalPassword(&password); got == nil || *got != password {
		t.Errorf("optionalPassword() = %v, want exact value %q", got, password)
	}
}

func TestRegisterUsernameValidation(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{name: "normalizes uppercase and whitespace", username: " Valery_Kirkizh ", wantErr: false},
		{name: "rejects invalid username", username: "valery__kirkizh", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			errs := validateRegisterRequest(registerRequest{
				Username:    test.username,
				Email:       "user@example.com",
				Password:    "secret",
				DisplayName: "Traveler",
			})

			_, gotErr := errs["username"]
			if gotErr != test.wantErr {
				t.Errorf("username error present = %t, want %t; errors: %v", gotErr, test.wantErr, errs)
			}
		})
	}
}

func TestIsValidInviteCode(t *testing.T) {
	tests := []struct {
		name      string
		submitted string
		expected  string
		want      bool
	}{
		{name: "matching code", submitted: "my-code", expected: "my-code", want: true},
		{name: "incorrect code", submitted: "wrong-code", expected: "my-code", want: false},
		{name: "missing submitted code", submitted: "", expected: "my-code", want: false},
		{name: "empty configured code", submitted: "my-code", expected: "", want: false},
		{name: "whitespace configured code", submitted: "my-code", expected: "   ", want: false},
		{name: "surrounding whitespace", submitted: "  my-code\t", expected: "\nmy-code ", want: true},
		{name: "case sensitive", submitted: "MY-CODE", expected: "my-code", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := isValidInviteCode(test.submitted, test.expected); got != test.want {
				t.Errorf("isValidInviteCode() = %t, want %t", got, test.want)
			}
		})
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
