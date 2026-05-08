package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrCurrentPasswordInvalid = errors.New("current password is invalid")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

type sessionCreator interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}

type UpdateProfileInput struct {
	UserID          string
	DisplayName     string
	Email           string
	CurrentPassword *string
	NewPassword     *string
}

func (r *Repository) Register(ctx context.Context, username, email, password, displayName string) (*User, string, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, "", err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var id uuid.UUID

	err = tx.QueryRow(ctx, `
		INSERT INTO users (username, email, password_hash, display_name)
		VALUES ($1, $2, $3, $4)
		RETURNING id, username, email, display_name
	`, username, email, string(passwordHash), displayName).Scan(
		&id,
		&username,
		&email,
		&displayName,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, "", ErrUserAlreadyExists
		}

		return nil, "", err
	}

	sessionToken, err := createSession(ctx, tx, id.String())
	if err != nil {
		return nil, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", err
	}

	user := NewUser(id, username, email, displayName)

	return &user, sessionToken, nil
}

func (r *Repository) Login(ctx context.Context, email, password string) (*User, string, error) {
	var (
		id           uuid.UUID
		username     string
		displayName  string
		passwordHash string
	)

	err := r.db.QueryRow(ctx, `
		SELECT id, username, email, password_hash, display_name
		FROM users
		WHERE email = $1
	`, email).Scan(
		&id,
		&username,
		&email,
		&passwordHash,
		&displayName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, "", ErrInvalidCredentials
	}
	if err != nil {
		return nil, "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	sessionToken, err := createSession(ctx, r.db, id.String())
	if err != nil {
		return nil, "", err
	}

	user := NewUser(id, username, email, displayName)

	return &user, sessionToken, nil
}

func (r *Repository) Logout(ctx context.Context, sessionToken string) error {
	tokenHash := hashToken(sessionToken)

	_, err := r.db.Exec(ctx, `
		DELETE FROM sessions
		WHERE token_hash = $1
	`, tokenHash)

	return err
}

func (r *Repository) CurrentUser(ctx context.Context, sessionToken string) (*User, error) {
	tokenHash := hashToken(sessionToken)

	var (
		id          uuid.UUID
		username    string
		email       string
		displayName string
	)

	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.username, u.email, u.display_name
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
		  AND s.expires_at > now()
	`, tokenHash).Scan(
		&id,
		&username,
		&email,
		&displayName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	user := NewUser(id, username, email, displayName)

	return &user, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, input UpdateProfileInput) (*User, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if input.NewPassword != nil {
		var passwordHash string
		err := tx.QueryRow(ctx, `
			SELECT password_hash
			FROM users
			WHERE id = $1
		`, input.UserID).Scan(&passwordHash)
		if err != nil {
			return nil, err
		}
		if input.CurrentPassword == nil ||
			bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(*input.CurrentPassword)) != nil {
			return nil, ErrCurrentPasswordInvalid
		}

		newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(*input.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(ctx, `
			UPDATE users
			SET password_hash = $1,
			    updated_at = now()
			WHERE id = $2
		`, string(newPasswordHash), input.UserID)
		if err != nil {
			return nil, err
		}
	}

	var (
		id          uuid.UUID
		username    string
		email       string
		displayName string
	)
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET display_name = $1,
		    email = $2,
		    updated_at = now()
		WHERE id = $3
		RETURNING id, username, email, display_name
	`, input.DisplayName, input.Email, input.UserID).Scan(
		&id,
		&username,
		&email,
		&displayName,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserAlreadyExists
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	user := NewUser(id, username, email, displayName)

	return &user, nil
}

func createSession(ctx context.Context, db sessionCreator, userID string) (string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}

	tokenHash := hashToken(token)
	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	_, err = db.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func randomToken(size int) (string, error) {
	bytes := make([]byte, size)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
