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
	var userID string

	response := &MapResponse{}

	err := r.db.QueryRow(ctx, `
		SELECT id, username, email, display_name
		FROM users
		WHERE username = $1
	`, username).Scan(
		&userID,
		&response.User.Username,
		&response.User.Email,
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

	flights, totalDistanceKM, flightHours, err := r.getFlights(ctx, userID)
	if err != nil {
		return nil, err
	}

	response.User.AvatarURL = auth.GravatarURL(response.User.Email)

	response.Places = places
	response.Flights = flights

	response.Stats = Stats{
		CountriesVisited: countUniqueCountries(places),
		PlacesVisited:    len(places),
		FlightsTaken:     len(flights),
		FlightDistanceKM: totalDistanceKM,
		FlightHours:      flightHours,
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

func (r *Repository) getFlights(ctx context.Context, userID string) ([]Flight, int, int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			f.id,
			f.from_airport_iata,
			f.to_airport_iata,
			f.departure_time,
			f.arrival_time,
			from_airport.lat,
			from_airport.lng,
			to_airport.lat,
			to_airport.lng,
			COALESCE(f.distance_km, 0)
		FROM flights f
		JOIN airports from_airport ON from_airport.iata_code = f.from_airport_iata
		JOIN airports to_airport ON to_airport.iata_code = f.to_airport_iata
		WHERE f.user_id = $1
		ORDER BY f.created_at ASC
	`, userID)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	flights := make([]Flight, 0)
	totalDistanceKM := 0
	totalFlightHours := 0

	for rows.Next() {
		var flight Flight
		var distanceKM int

		if err := rows.Scan(
			&flight.ID,
			&flight.From,
			&flight.To,
			&flight.DepartureTime,
			&flight.ArrivalTime,
			&flight.FromPoint.Lat,
			&flight.FromPoint.Lng,
			&flight.ToPoint.Lat,
			&flight.ToPoint.Lng,
			&distanceKM,
		); err != nil {
			return nil, 0, 0, err
		}

		totalDistanceKM += distanceKM

		if flight.DepartureTime != nil && flight.ArrivalTime != nil {
			totalFlightHours += int(flight.ArrivalTime.Sub(*flight.DepartureTime).Hours())
		}

		flights = append(flights, flight)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, 0, err
	}

	return flights, totalDistanceKM, totalFlightHours, nil
}

func countUniqueCountries(places []Place) int {
	countries := make(map[string]struct{})

	for _, place := range places {
		countries[place.CountryCode] = struct{}{}
	}

	return len(countries)
}
