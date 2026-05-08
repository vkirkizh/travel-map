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
	"strings"
	"time"

	"github.com/vkirkizh/travel-map/backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type countryCSVRow struct {
	Code string
	Name string
}

func main() {
	_ = godotenv.Load()

	filePath := flag.String("file", "", "Path to countries.csv")
	flag.Parse()

	if strings.TrimSpace(*filePath) == "" {
		slog.Error("missing required -file argument")
		os.Exit(1)
	}

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	imported, skipped, err := importCountries(ctx, db, *filePath)
	if err != nil {
		slog.Error("failed to import countries", "error", err)
		os.Exit(1)
	}

	slog.Info("countries import completed", "imported", imported, "skipped", skipped)
}

func importCountries(ctx context.Context, db *pgxpool.Pool, filePath string) (int, int, error) {
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

	requiredColumns := []string{"code", "name"}
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

		row, ok := parseCountryRow(record, indexes)
		if !ok {
			skipped++
			continue
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO countries (code, name)
			VALUES ($1, $2)
			ON CONFLICT (code) DO UPDATE SET
				name = EXCLUDED.name
		`, row.Code, row.Name)
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

func parseCountryRow(record []string, indexes map[string]int) (countryCSVRow, bool) {
	code := strings.ToUpper(strings.TrimSpace(getCSVValue(record, indexes, "code")))
	name := strings.TrimSpace(getCSVValue(record, indexes, "name"))

	if len(code) != 2 || name == "" {
		return countryCSVRow{}, false
	}

	return countryCSVRow{
		Code: code,
		Name: name,
	}, true
}

func getCSVValue(record []string, indexes map[string]int, column string) string {
	index, ok := indexes[column]
	if !ok || index >= len(record) {
		return ""
	}

	return record[index]
}
