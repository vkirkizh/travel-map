package publicmap

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vkirkizh/travel-map/backend/internal/auth"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*MapResponse, error) {
	var (
		userID string
		email  string
	)

	response := &MapResponse{}

	err := r.db.QueryRow(ctx, `
		SELECT id, username, email, display_name
		FROM users
		WHERE username = $1
	`, username).Scan(
		&userID,
		&response.User.Username,
		&email,
		&response.User.DisplayName,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	places, err := r.getPlaces(ctx, userID)
	if err != nil {
		return nil, err
	}

	response.User.AvatarURL = auth.GravatarURL(email)

	response.Places = places

	response.Stats = Stats{
		CountriesVisited: countUniqueCountries(places),
		PlacesVisited:    len(places),
	}

	return response, nil
}

func (r *Repository) getPlaces(ctx context.Context, userID string) ([]Place, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, country_code, lat, lng
		FROM places
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	places := make([]Place, 0)

	for rows.Next() {
		var place Place

		if err := rows.Scan(
			&place.ID,
			&place.Title,
			&place.CountryCode,
			&place.Lat,
			&place.Lng,
		); err != nil {
			return nil, err
		}

		places = append(places, place)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return places, nil
}

func countUniqueCountries(places []Place) int {
	countries := make(map[string]struct{})

	for _, place := range places {
		countries[place.CountryCode] = struct{}{}
	}

	return len(countries)
}
