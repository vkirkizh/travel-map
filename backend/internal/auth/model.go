package auth

import "github.com/google/uuid"

type User struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
}

func NewUser(id uuid.UUID, username, email, displayName string) User {
	return User{
		ID:          id,
		Username:    username,
		Email:       email,
		DisplayName: displayName,
		AvatarURL:   GravatarURL(email),
	}
}
