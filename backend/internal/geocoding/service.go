package geocoding

import (
	"context"
	"encoding/json"
	"errors"
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

	result, err := s.resolveViaNominatim(ctx, query, normalized)
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
	DisplayName string `json:"display_name"`
	Lat         string `json:"lat"`
	Lon         string `json:"lon"`
	Address     struct {
		Country     string `json:"country"`
		CountryCode string `json:"country_code"`
	} `json:"address"`
}

func (s *Service) resolveViaNominatim(ctx context.Context, originalQuery string, normalized string) (*Result, error) {
	endpoint, err := url.Parse(s.baseURL + "/search")
	if err != nil {
		return nil, err
	}

	params := endpoint.Query()
	params.Set("q", normalized)
	params.Set("format", "jsonv2")
	params.Set("addressdetails", "1")
	params.Set("limit", "1")
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

	var items []nominatimResponseItem
	if err := json.NewDecoder(response.Body).Decode(&items); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return nil, ErrNotFound
	}

	item := items[0]

	lat, err := strconv.ParseFloat(item.Lat, 64)
	if err != nil {
		return nil, err
	}

	lng, err := strconv.ParseFloat(item.Lon, 64)
	if err != nil {
		return nil, err
	}

	rawJSON, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	countryCode := strings.ToUpper(item.Address.CountryCode)
	if countryCode == "" {
		return nil, ErrNotFound
	}

	return &Result{
		Title:       item.DisplayName,
		Query:       originalQuery,
		CountryCode: countryCode,
		CountryName: item.Address.Country,
		Lat:         lat,
		Lng:         lng,
		RawJSON:     rawJSON,
	}, nil
}
