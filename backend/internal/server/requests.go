package server

import (
	"crypto/subtle"
	"strings"
	"unicode/utf8"

	"github.com/vkirkizh/travel-map/backend/internal/auth"
)

type registerRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	InviteCode  string `json:"invite_code"`
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

func validateRegisterRequest(request registerRequest) map[string]string {
	errs := make(map[string]string)

	username := auth.NormalizeUsername(request.Username)
	email := auth.NormalizeEmail(request.Email)

	if username == "" {
		errs["username"] = "Username is required."
	} else if !auth.IsValidUsername(username) {
		errs["username"] = "Username is invalid."
	}

	if email == "" {
		errs["email"] = "Email is required."
	} else if !auth.IsValidEmail(email) {
		errs["email"] = "Email is invalid."
	}

	if message := validateNewPassword(request.Password); message != "" {
		errs["password"] = message
	}

	if message := validateDisplayName(request.DisplayName); message != "" {
		errs["display_name"] = message
	}

	return errs
}

func isValidInviteCode(submitted string, expected string) bool {
	submitted = strings.TrimSpace(submitted)
	expected = strings.TrimSpace(expected)

	if submitted == "" || expected == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(submitted), []byte(expected)) == 1
}

func validateLoginRequest(request loginRequest) map[string]string {
	errs := make(map[string]string)

	email := auth.NormalizeEmail(request.Email)
	password := request.Password

	if email == "" {
		errs["email"] = "Email is required."
	} else if !auth.IsValidEmail(email) {
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

	if message := validateDisplayName(request.DisplayName); message != "" {
		errs["display_name"] = message
	}

	if email == "" {
		errs["email"] = "Email is required."
	} else if !auth.IsValidEmail(email) {
		errs["email"] = "Email is invalid."
	}

	newPassword := optionalPassword(request.NewPassword)
	currentPassword := optionalPassword(request.CurrentPassword)

	if newPassword != nil {
		if message := validateNewPassword(*newPassword); message != "" {
			errs["new_password"] = message
		}

		if currentPassword == nil {
			errs["current_password"] = "Current password is required to change password."
		}
	}

	return errs
}

func validateCreatePlaceRequest(request createPlaceRequest) map[string]string {
	errs := make(map[string]string)
	queryLength := utf8.RuneCountInString(strings.TrimSpace(request.Query))

	if queryLength < 3 {
		errs["query"] = "Place query must be at least 3 characters."
	} else if queryLength > 200 {
		errs["query"] = "Place query must be at most 200 characters."
	}

	return errs
}

func validateDisplayName(displayName string) string {
	length := utf8.RuneCountInString(strings.TrimSpace(displayName))

	if length < 2 {
		return "Display name must be at least 2 characters."
	}
	if length > 50 {
		return "Display name must be at most 50 characters."
	}

	return ""
}

func validateNewPassword(password string) string {
	if len(password) < 6 {
		return "Password must be at least 6 bytes."
	}
	if len(password) > 64 {
		return "Password must be at most 64 bytes."
	}

	return ""
}
