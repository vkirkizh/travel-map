package publicmap

import "github.com/google/uuid"

type User struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type Place struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	CountryCode string    `json:"country_code"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
}

type Stats struct {
	CountriesVisited int `json:"countries_visited"`
	PlacesVisited    int `json:"places_visited"`
}

type MapResponse struct {
	User   User    `json:"user"`
	Places []Place `json:"places"`
	Stats  Stats   `json:"stats"`
}
