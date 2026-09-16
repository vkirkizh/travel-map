package geocoding

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("geocoding result not found")

type Service struct {
	db        *pgxpool.Pool
	baseURL   string
	userAgent string
	client    *http.Client
}

func NewService(db *pgxpool.Pool, baseURL string, userAgent string) *Service {
	return &Service{
		db:        db,
		baseURL:   strings.TrimRight(baseURL, "/"),
		userAgent: userAgent,
		client: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

func (s *Service) Resolve(ctx context.Context, query string) (*Result, error) {
	normalized := normalizeQuery(query)

	if normalized == "" {
		return nil, ErrNotFound
	}

	cached, err := s.getFromCache(ctx, normalized)
	if err == nil {
		return cached, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}

	result, err := s.resolveViaNominatim(ctx, query)
	if err != nil {
		return nil, err
	}

	if err := s.saveToCache(ctx, normalized, result); err != nil {
		return nil, err
	}

	return result, nil
}

func normalizeQuery(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(query))), " ")
}

type nominatimResponseItem struct {
	Name        string            `json:"name"`
	NameDetails map[string]string `json:"namedetails"`
	Lat         string            `json:"lat"`
	Lon         string            `json:"lon"`
	Address     struct {
		Country       string `json:"country"`
		CountryCode   string `json:"country_code"`
		ISO3166Level3 string `json:"ISO3166-2-lvl3"`
	} `json:"address"`
}

func (s *Service) resolveViaNominatim(ctx context.Context, originalQuery string) (*Result, error) {
	endpoint, err := url.Parse(s.baseURL + "/search")
	if err != nil {
		return nil, err
	}

	params := endpoint.Query()
	params.Set("q", originalQuery)
	params.Set("format", "jsonv2")
	params.Set("addressdetails", "1")
	params.Set("namedetails", "1")
	params.Set("accept-language", "en")
	params.Set("limit", "5")
	endpoint.RawQuery = params.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("User-Agent", s.userAgent)
	request.Header.Set("Accept", "application/json")

	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, errors.New("nominatim request failed")
	}

	return parseNominatimResponse(response.Body, originalQuery)
}

func parseNominatimResponse(reader io.Reader, query string) (*Result, error) {
	var items []nominatimResponseItem
	if err := json.NewDecoder(reader).Decode(&items); err != nil {
		return nil, err
	}

	for _, item := range items {
		result, err := normalizeNominatimItem(item, query)
		if err == nil {
			return result, nil
		}
		if !errors.Is(err, ErrNotFound) {
			return nil, err
		}
	}

	return nil, ErrNotFound
}

func normalizeNominatimItem(item nominatimResponseItem, query string) (*Result, error) {
	placeName := ""
	for _, candidate := range []string{
		item.NameDetails["_place_name:en"],
		item.NameDetails["name:en"],
		item.Name,
	} {
		if candidate = strings.TrimSpace(candidate); candidate != "" {
			placeName = candidate
			break
		}
	}

	countryName := strings.TrimSpace(item.Address.Country)
	countryCode := strings.ToUpper(strings.TrimSpace(item.Address.CountryCode))
	switch countryName {
	case "Abkhazia":
		countryCode = "AB"
		countryName = "Abkhazia"
	case "South Ossetia":
		countryCode = "OS"
		countryName = "South Ossetia"
	case "Northern Cyprus":
		countryCode = "NC"
		countryName = "Northern Cyprus"
	default:
		switch strings.TrimSpace(item.Address.ISO3166Level3) {
		case "CN-HK":
			countryCode = "HK"
			countryName = "Hong Kong"
		case "CN-MO":
			countryCode = "MO"
			countryName = "Macau"
		}
	}
	if placeName == "" || countryName == "" || countryCode == "" {
		return nil, ErrNotFound
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(item.Lat), 64)
	if err != nil || math.IsNaN(lat) || math.IsInf(lat, 0) || lat < -90 || lat > 90 {
		return nil, ErrNotFound
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(item.Lon), 64)
	if err != nil || math.IsNaN(lng) || math.IsInf(lng, 0) || lng < -180 || lng > 180 {
		return nil, ErrNotFound
	}

	rawJSON, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	return &Result{
		Title:       placeName + ", " + countryName,
		Query:       query,
		CountryCode: countryCode,
		CountryName: countryName,
		Lat:         lat,
		Lng:         lng,
		RawJSON:     rawJSON,
	}, nil
}
