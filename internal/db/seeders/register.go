package seeders

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type seedFunc func(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error

type seedFile struct {
	Name      string
	Timestamp time.Time
	Up        seedFunc
	Down      seedFunc
}

var seedRegistry = make(map[string]*seedFile)

func registerSeed(up seedFunc, down seedFunc) error {
	_, filename, _, ok := runtime.Caller(1)
	if !ok {
		return fmt.Errorf("failed to get caller information")
	}

	name, timestamp, err := parseSeedFilename(filename)
	if err != nil {
		return fmt.Errorf("failed to parse seed filename %s: %w", filename, err)
	}

	seedRegistry[name] = &seedFile{
		Name:      name,
		Timestamp: timestamp,
		Up:        up,
		Down:      down,
	}
	return nil
}

func LoadSeeds() ([]db.Seed, error) {
	var seeds []db.Seed

	for _, seedFile := range seedRegistry {
		seeds = append(seeds, db.Seed{
			Name:      seedFile.Name,
			Timestamp: seedFile.Timestamp,
			Up:        seedFile.Up,
			Down:      seedFile.Down,
		})
	}

	sort.Slice(seeds, func(i, j int) bool {
		return seeds[i].Timestamp.Before(seeds[j].Timestamp)
	})

	return seeds, nil
}

func parseSeedFilename(filename string) (string, time.Time, error) {
	base := filepath.Base(filename)
	if !strings.HasSuffix(base, ".go") {
		return "", time.Time{}, fmt.Errorf("not a Go file: %s", filename)
	}

	base = strings.TrimSuffix(base, ".go")

	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return "", time.Time{}, fmt.Errorf("invalid seed filename format: %s (expected: YYYYMMDDHHMMSS_seed_name.go)", filename)
	}

	timestampStr := parts[0]
	if len(timestampStr) != 14 {
		return "", time.Time{}, fmt.Errorf("invalid timestamp format: %s (expected: YYYYMMDDHHMMSS)", timestampStr)
	}

	timestamp, err := time.Parse("20060102150405", timestampStr)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	seedName := base
	return seedName, timestamp, nil
}

func GetSeedersDir() string {
	_, dbFile, _, _ := runtime.Caller(0)
	dbBasepath := filepath.Dir(dbFile)
	return filepath.Join(dbBasepath, "seeders")
}
