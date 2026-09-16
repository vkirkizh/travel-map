package geocoding

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestNewServiceConfiguresNominatimTimeout(t *testing.T) {
	service := NewService(nil)

	if service.client.Timeout != 10*time.Second {
		t.Errorf("client timeout = %s, want %s", service.client.Timeout, 10*time.Second)
	}
}

func TestParseNominatimResponseValidResult(t *testing.T) {
	response := `[
		{
			"name": "Berlin",
			"display_name": "Berlin, Deutschland, Europe",
			"lat": "52.5173885",
			"lon": "13.3951309",
			"address": {"country": "Germany", "country_code": "de"}
		}
	]`

	result, err := parseNominatimResponse(strings.NewReader(response), "Berlin")
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}

	if result.Title != "Berlin, Germany" {
		t.Errorf("Title = %q, want %q", result.Title, "Berlin, Germany")
	}
	if result.CountryCode != "DE" {
		t.Errorf("CountryCode = %q, want %q", result.CountryCode, "DE")
	}
	if result.CountryName != "Germany" {
		t.Errorf("CountryName = %q, want %q", result.CountryName, "Germany")
	}
	if result.Lat != 52.5173885 || result.Lng != 13.3951309 {
		t.Errorf("coordinates = (%v, %v), want (%v, %v)", result.Lat, result.Lng, 52.5173885, 13.3951309)
	}
}

func TestParseNominatimResponseNormalizesCountryCode(t *testing.T) {
	response := `[
		{
			"name": "Luxembourg",
			"lat": "49.6112768",
			"lon": "6.129799",
			"address": {"country": "Luxembourg", "country_code": "lu"}
		}
	]`

	result, err := parseNominatimResponse(strings.NewReader(response), "Luxembourg")
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}

	if result.CountryCode != "LU" {
		t.Errorf("CountryCode = %q, want %q", result.CountryCode, "LU")
	}
}

func TestParseNominatimResponseUsesPlaceNameFallbackPriority(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
	}{
		{
			name: "place name before English and top-level names",
			response: `[
				{
					"name": "City of Zagreb",
					"namedetails": {
						"_place_name:en": "  Zagreb  ",
						"name:en": "City of Zagreb"
					},
					"lat": "45.8130967",
					"lon": "15.9772795",
					"address": {
						"city": "City of Zagreb",
						"country": "Croatia",
						"country_code": "hr"
					}
				}
			]`,
			want: "Zagreb, Croatia",
		},
		{
			name: "English name before top-level name",
			response: `[
				{
					"name": "Greater London",
					"namedetails": {
						"_place_name:en": "  ",
						"name:en": "  London  "
					},
					"lat": "51.5074456",
					"lon": "-0.1277653",
					"address": {
						"city": "Greater London",
						"country": "United Kingdom",
						"country_code": "gb"
					}
				}
			]`,
			want: "London, United Kingdom",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := parseNominatimResponse(strings.NewReader(test.response), test.name)
			if err != nil {
				t.Fatalf("parse response: %v", err)
			}

			if result.Title != test.want {
				t.Errorf("Title = %q, want %q", result.Title, test.want)
			}
		})
	}
}

func TestNormalizeNominatimItemTravelDestinationOverrides(t *testing.T) {
	tests := []struct {
		name          string
		placeName     string
		country       string
		countryCode   string
		iso3166Level3 string
		wantCode      string
		wantName      string
	}{
		{
			name:        "Abkhazia",
			placeName:   "Sukhumi",
			country:     "Abkhazia",
			countryCode: "ge",
			wantCode:    "AB",
			wantName:    "Abkhazia",
		},
		{
			name:        "South Ossetia",
			placeName:   "Tskhinvali",
			country:     "South Ossetia",
			countryCode: "ge",
			wantCode:    "OS",
			wantName:    "South Ossetia",
		},
		{
			name:        "Northern Cyprus",
			placeName:   "Kyrenia",
			country:     "Northern Cyprus",
			countryCode: "cy",
			wantCode:    "NC",
			wantName:    "Northern Cyprus",
		},
		{
			name:          "Hong Kong",
			placeName:     "Hong Kong",
			country:       "China",
			countryCode:   "cn",
			iso3166Level3: "CN-HK",
			wantCode:      "HK",
			wantName:      "Hong Kong",
		},
		{
			name:          "Macau",
			placeName:     "Macau",
			country:       "China",
			countryCode:   "cn",
			iso3166Level3: "CN-MO",
			wantCode:      "MO",
			wantName:      "Macau",
		},
		{
			name:          "Georgia is not overridden",
			placeName:     "Tbilisi",
			country:       "Georgia",
			countryCode:   "ge",
			iso3166Level3: "GE-TB",
			wantCode:      "GE",
			wantName:      "Georgia",
		},
		{
			name:          "Cyprus is not overridden",
			placeName:     "Nicosia",
			country:       "Cyprus",
			countryCode:   "cy",
			iso3166Level3: "CY-01",
			wantCode:      "CY",
			wantName:      "Cyprus",
		},
		{
			name:          "Kosovo is not overridden",
			placeName:     "Pristina",
			country:       "Kosovo",
			countryCode:   "xk",
			iso3166Level3: "XK-01",
			wantCode:      "XK",
			wantName:      "Kosovo",
		},
		{
			name:          "Palestinian Territories is not overridden",
			placeName:     "Ramallah",
			country:       "Palestinian Territories",
			countryCode:   "ps",
			iso3166Level3: "PS-RBH",
			wantCode:      "PS",
			wantName:      "Palestinian Territories",
		},
		{
			name:          "Taiwan is not overridden",
			placeName:     "Taipei",
			country:       "Taiwan",
			countryCode:   "tw",
			iso3166Level3: "TW-TPE",
			wantCode:      "TW",
			wantName:      "Taiwan",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			item := nominatimResponseItem{
				Name: test.placeName,
				Lat:  "1",
				Lon:  "2",
			}
			item.Address.Country = test.country
			item.Address.CountryCode = test.countryCode
			item.Address.ISO3166Level3 = test.iso3166Level3

			result, err := normalizeNominatimItem(item, test.placeName)
			if err != nil {
				t.Fatalf("normalize item: %v", err)
			}

			if result.CountryCode != test.wantCode {
				t.Errorf("CountryCode = %q, want %q", result.CountryCode, test.wantCode)
			}
			if result.CountryName != test.wantName {
				t.Errorf("CountryName = %q, want %q", result.CountryName, test.wantName)
			}
			wantTitle := test.placeName + ", " + test.wantName
			if result.Title != wantTitle {
				t.Errorf("Title = %q, want %q", result.Title, wantTitle)
			}
		})
	}
}

func TestParseNominatimResponseIgnoresIncompleteResult(t *testing.T) {
	response := `[
		{
			"display_name": "Stonehenge, Amesbury, Wiltshire, England, United Kingdom",
			"lat": "51.1788853",
			"lon": "-1.8262144",
			"address": {"country": "United Kingdom", "country_code": "gb"}
		},
		{
			"name": "Stonehenge",
			"lat": "51.1788853",
			"lon": "-1.8262144",
			"address": {"country": "United Kingdom", "country_code": "gb"}
		}
	]`

	result, err := parseNominatimResponse(strings.NewReader(response), "Stonehenge")
	if err != nil {
		t.Fatalf("parse response: %v", err)
	}

	if result.Title != "Stonehenge, United Kingdom" {
		t.Errorf("Title = %q, want %q", result.Title, "Stonehenge, United Kingdom")
	}
}

func TestParseNominatimResponseRejectsInvalidCoordinates(t *testing.T) {
	response := `[
		{
			"name": "Invalid",
			"lat": "91",
			"lon": "0",
			"address": {"country": "Nowhere", "country_code": "xx"}
		}
	]`

	_, err := parseNominatimResponse(strings.NewReader(response), "Invalid")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v, want ErrNotFound", err)
	}
}

func TestResolveViaNominatimSendsRequiredSearchParameters(t *testing.T) {
	service := NewService(nil)
	service.client.Transport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		query := request.URL.Query()
		want := map[string]string{
			"q":               "Berlin Germany",
			"format":          "jsonv2",
			"addressdetails":  "1",
			"namedetails":     "1",
			"accept-language": "en",
			"limit":           "5",
		}
		for key, wantValue := range want {
			if got := query.Get(key); got != wantValue {
				t.Errorf("query parameter %q = %q, want %q", key, got, wantValue)
			}
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(`[
			{
				"name": "Berlin",
				"lat": "52.5173885",
				"lon": "13.3951309",
				"address": {"country": "Germany", "country_code": "de"}
			}
		]`)),
			Header: make(http.Header),
		}, nil
	})

	_, err := service.resolveViaNominatim(context.Background(), "Berlin Germany")
	if err != nil {
		t.Fatalf("resolve via Nominatim: %v", err)
	}
}
