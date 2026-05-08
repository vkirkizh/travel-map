package publicmap

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Username    string  `json:"username"`
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
}

type Place struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	CountryCode string    `json:"country_code"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Flight struct {
	ID            uuid.UUID  `json:"id"`
	From          string     `json:"from"`
	To            string     `json:"to"`
	FromPoint     Point      `json:"from_point"`
	ToPoint       Point      `json:"to_point"`
	DepartureTime *time.Time `json:"departure_time"`
	ArrivalTime   *time.Time `json:"arrival_time"`
}

type Stats struct {
	CountriesVisited int `json:"countries_visited"`
	PlacesVisited    int `json:"places_visited"`
	FlightsTaken     int `json:"flights_taken"`
	FlightDistanceKM int `json:"flight_distance_km"`
	FlightHours      int `json:"flight_hours"`
}

type MapResponse struct {
	User    User     `json:"user"`
	Places  []Place  `json:"places"`
	Flights []Flight `json:"flights"`
	Stats   Stats    `json:"stats"`
}
