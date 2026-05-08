package places

import "github.com/google/uuid"

type Place struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Query       string    `json:"query"`
	CountryCode string    `json:"country_code"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
}
