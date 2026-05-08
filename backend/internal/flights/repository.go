package flights

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAirportNotFound = errors.New("airport not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]Flight, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			f.id,
			f.from_airport_iata,
			f.to_airport_iata,
			f.departure_time,
			f.arrival_time,
			f.flight_number,
			COALESCE(f.distance_km, 0),
			from_airport.lat,
			from_airport.lng,
			to_airport.lat,
			to_airport.lng
		FROM flights f
		JOIN airports from_airport ON from_airport.iata_code = f.from_airport_iata
		JOIN airports to_airport ON to_airport.iata_code = f.to_airport_iata
		WHERE f.user_id = $1
		ORDER BY f.departure_time NULLS LAST, f.created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Flight, 0)

	for rows.Next() {
		var flight Flight

		if err := rows.Scan(
			&flight.ID,
			&flight.FromAirportIATA,
			&flight.ToAirportIATA,
			&flight.DepartureTime,
			&flight.ArrivalTime,
			&flight.FlightNumber,
			&flight.DistanceKM,
			&flight.FromPoint.Lat,
			&flight.FromPoint.Lng,
			&flight.ToPoint.Lat,
			&flight.ToPoint.Lng,
		); err != nil {
			return nil, err
		}

		result = append(result, flight)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) Create(
	ctx context.Context,
	userID uuid.UUID,
	fromAirportIATA string,
	toAirportIATA string,
	departureTime *time.Time,
	arrivalTime *time.Time,
	flightNumber *string,
) (*Flight, error) {
	fromAirportIATA = strings.ToUpper(strings.TrimSpace(fromAirportIATA))
	toAirportIATA = strings.ToUpper(strings.TrimSpace(toAirportIATA))

	fromPoint, err := r.getAirportPoint(ctx, fromAirportIATA)
	if err != nil {
		return nil, err
	}

	toPoint, err := r.getAirportPoint(ctx, toAirportIATA)
	if err != nil {
		return nil, err
	}

	distanceKM := DistanceKM(fromPoint, toPoint)

	var created Flight

	err = r.db.QueryRow(ctx, `
		INSERT INTO flights (
			user_id,
			from_airport_iata,
			to_airport_iata,
			departure_time,
			arrival_time,
			flight_number,
			distance_km
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			from_airport_iata,
			to_airport_iata,
			departure_time,
			arrival_time,
			flight_number,
			distance_km
	`,
		userID,
		fromAirportIATA,
		toAirportIATA,
		departureTime,
		arrivalTime,
		flightNumber,
		distanceKM,
	).Scan(
		&created.ID,
		&created.FromAirportIATA,
		&created.ToAirportIATA,
		&created.DepartureTime,
		&created.ArrivalTime,
		&created.FlightNumber,
		&created.DistanceKM,
	)
	if err != nil {
		return nil, err
	}

	created.FromPoint = fromPoint
	created.ToPoint = toPoint

	return &created, nil
}

func (r *Repository) Delete(ctx context.Context, userID uuid.UUID, flightID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM flights
		WHERE id = $1
		  AND user_id = $2
	`, flightID, userID)

	return err
}

func (r *Repository) getAirportPoint(ctx context.Context, iataCode string) (Point, error) {
	var point Point

	err := r.db.QueryRow(ctx, `
		SELECT lat, lng
		FROM airports
		WHERE iata_code = $1
	`, iataCode).Scan(&point.Lat, &point.Lng)
	if errors.Is(err, pgx.ErrNoRows) {
		return Point{}, ErrAirportNotFound
	}
	if err != nil {
		return Point{}, err
	}

	return point, nil
}
