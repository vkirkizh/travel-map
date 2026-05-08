package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vkirkizh/travel-map/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type airportCSVRow struct {
	Type        string
	Name        string
	Lat         float64
	Lng         float64
	CountryCode string
	City        string
	IATACode    string
}

func main() {
	_ = godotenv.Load()

	filePath := flag.String("file", "", "Path to airports.csv")
	flag.Parse()

	if strings.TrimSpace(*filePath) == "" {
		slog.Error("missing required -file argument")
		os.Exit(1)
	}

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	imported, skipped, err := importAirports(ctx, db, *filePath)
	if err != nil {
		slog.Error("failed to import airports", "error", err)
		os.Exit(1)
	}

	slog.Info("airports import completed", "imported", imported, "skipped", skipped)
}

func importAirports(ctx context.Context, db *pgxpool.Pool, filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = file.Close()
	}()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return 0, 0, err
	}

	indexes := buildHeaderIndex(header)

	requiredColumns := []string{
		"type",
		"name",
		"latitude_deg",
		"longitude_deg",
		"iso_country",
		"municipality",
		"iata_code",
	}

	for _, column := range requiredColumns {
		if _, ok := indexes[column]; !ok {
			return 0, 0, fmt.Errorf("required CSV column is missing: %s", column)
		}
	}

	tx, err := db.Begin(ctx)
	if err != nil {
		return 0, 0, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	imported := 0
	skipped := 0

	for {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return imported, skipped, err
		}

		row, ok := parseAirportRow(record, indexes)
		if !ok {
			skipped++
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO countries (code, name)
			VALUES ($1, $2)
			ON CONFLICT (code) DO NOTHING
		`, row.CountryCode, row.CountryCode)
		if err != nil {
			return imported, skipped, err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO airports (
				iata_code,
				name,
				city,
				country_code,
				lat,
				lng
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (iata_code) DO UPDATE SET
				name = EXCLUDED.name,
				city = EXCLUDED.city,
				country_code = EXCLUDED.country_code,
				lat = EXCLUDED.lat,
				lng = EXCLUDED.lng
		`,
			row.IATACode,
			row.Name,
			nullableString(row.City),
			row.CountryCode,
			row.Lat,
			row.Lng,
		)
		if err != nil {
			return imported, skipped, err
		}

		imported++
	}

	if err := tx.Commit(ctx); err != nil {
		return imported, skipped, err
	}

	return imported, skipped, nil
}

func buildHeaderIndex(header []string) map[string]int {
	indexes := make(map[string]int, len(header))

	for index, column := range header {
		indexes[strings.TrimSpace(column)] = index
	}

	return indexes
}

func parseAirportRow(record []string, indexes map[string]int) (airportCSVRow, bool) {
	airportType := getCSVValue(record, indexes, "type")
	if !isRealAirportType(airportType) {
		return airportCSVRow{}, false
	}

	iataCode := strings.ToUpper(strings.TrimSpace(getCSVValue(record, indexes, "iata_code")))
	if len(iataCode) != 3 {
		return airportCSVRow{}, false
	}

	name := strings.TrimSpace(getCSVValue(record, indexes, "name"))
	if name == "" {
		return airportCSVRow{}, false
	}

	countryCode := strings.ToUpper(strings.TrimSpace(getCSVValue(record, indexes, "iso_country")))
	if len(countryCode) != 2 {
		return airportCSVRow{}, false
	}

	lat, err := strconv.ParseFloat(strings.TrimSpace(getCSVValue(record, indexes, "latitude_deg")), 64)
	if err != nil {
		return airportCSVRow{}, false
	}

	lng, err := strconv.ParseFloat(strings.TrimSpace(getCSVValue(record, indexes, "longitude_deg")), 64)
	if err != nil {
		return airportCSVRow{}, false
	}

	return airportCSVRow{
		Type:        airportType,
		Name:        name,
		Lat:         lat,
		Lng:         lng,
		CountryCode: countryCode,
		City:        strings.TrimSpace(getCSVValue(record, indexes, "municipality")),
		IATACode:    iataCode,
	}, true
}

func getCSVValue(record []string, indexes map[string]int, column string) string {
	index, ok := indexes[column]
	if !ok || index >= len(record) {
		return ""
	}

	return record[index]
}

func isRealAirportType(value string) bool {
	switch strings.TrimSpace(value) {
	case "large_airport", "medium_airport", "small_airport":
		return true
	default:
		return false
	}
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	return value
}
