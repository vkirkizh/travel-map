package places

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]Place, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, title, query, country_code, lat, lng
		FROM places
		WHERE user_id = $1
		ORDER BY created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Place, 0)

	for rows.Next() {
		var place Place

		if err := rows.Scan(
			&place.ID,
			&place.Title,
			&place.Query,
			&place.CountryCode,
			&place.Lat,
			&place.Lng,
		); err != nil {
			return nil, err
		}

		result = append(result, place)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, place Place) (*Place, error) {
	var created Place

	err := r.db.QueryRow(ctx, `
		INSERT INTO places (user_id, country_code, title, query, lat, lng)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, query, country_code, lat, lng
	`,
		userID,
		place.CountryCode,
		place.Title,
		place.Query,
		place.Lat,
		place.Lng,
	).Scan(
		&created.ID,
		&created.Title,
		&created.Query,
		&created.CountryCode,
		&created.Lat,
		&created.Lng,
	)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (r *Repository) Delete(ctx context.Context, userID uuid.UUID, placeID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM places
		WHERE id = $1
		  AND user_id = $2
	`, placeID, userID)

	return err
}
