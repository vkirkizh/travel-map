package geocoding

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (s *Service) getFromCache(ctx context.Context, normalized string) (*Result, error) {
	var result Result

	err := s.db.QueryRow(ctx, `
		SELECT
			query_normalized,
			result_json,
			country_code,
			lat,
			lng
		FROM geocoding_cache
		WHERE query_normalized = $1
	`, normalized).Scan(
		&result.Query,
		&result.RawJSON,
		&result.CountryCode,
		&result.Lat,
		&result.Lng,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	var cached nominatimResponseItem
	if err := json.Unmarshal(result.RawJSON, &cached); err == nil {
		result.Title = cached.DisplayName
		result.CountryName = cached.Address.Country
	}

	return &result, nil
}

func (s *Service) saveToCache(ctx context.Context, normalized string, result *Result) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if result.CountryCode != "" && result.CountryName != "" {
		_, err = tx.Exec(ctx, `
			INSERT INTO countries (code, name)
			VALUES ($1, $2)
			ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		`, result.CountryCode, result.CountryName)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO geocoding_cache (
			query_normalized,
			provider,
			result_json,
			country_code,
			lat,
			lng
		)
		VALUES ($1, 'nominatim', $2, $3, $4, $5)
		ON CONFLICT (query_normalized) DO UPDATE SET
			result_json = EXCLUDED.result_json,
			country_code = EXCLUDED.country_code,
			lat = EXCLUDED.lat,
			lng = EXCLUDED.lng
	`, normalized, result.RawJSON, result.CountryCode, result.Lat, result.Lng)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
