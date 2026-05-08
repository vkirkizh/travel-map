package geocoding

type Result struct {
	Title       string
	Query       string
	CountryCode string
	CountryName string
	Lat         float64
	Lng         float64
	RawJSON     []byte
}
