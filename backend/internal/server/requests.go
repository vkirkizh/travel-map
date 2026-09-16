package server

import (
	"net/mail"
	"strings"

	"github.com/vkirkizh/travel-map/backend/internal/auth"
)

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
