package flights

import (
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	ID              uuid.UUID  `json:"id"`
	FromAirportIATA string     `json:"from_airport_iata"`
	ToAirportIATA   string     `json:"to_airport_iata"`
	DepartureTime   *time.Time `json:"departure_time"`
	ArrivalTime     *time.Time `json:"arrival_time"`
	FlightNumber    *string    `json:"flight_number"`
	DistanceKM      int        `json:"distance_km"`

	FromPoint Point `json:"from_point"`
	ToPoint   Point `json:"to_point"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}
