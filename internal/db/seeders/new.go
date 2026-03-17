package seeders

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/guncv/ticket-reservation-server/internal/config"
)

func NewSeeder(name string, cfg *config.Config) error {
	_, dbFile, _, _ := runtime.Caller(0)
	dbBasepath := filepath.Dir(dbFile)

	lowerName := strings.ToLower(name)
	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_seed_%s.go", timestamp, lowerName)
	seederPath := filepath.Join(dbBasepath, filename)

	if _, err := os.Stat(seederPath); !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to check seeder file: %w", err)
	}

	funcName := toCamelCase("seed_" + name)
	funcNameUp := funcName + "Up"
	funcNameDown := funcName + "Down"

	template := fmt.Sprintf(`package seeders
import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/config"
	"github.com/guncv/ticket-reservation-server/internal/containers"
	"github.com/jackc/pgx/v5/pgxpool"
)
	
func init() {
	if err := registerSeed(%s, %s); err != nil {
		panic(err)
	}
}
	
func %s(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	c := containers.NewContainer(cfg)
	
	if err := c.Container.Invoke(func() {
		// TODO: Implement seeder logic here
	}); err != nil {
		return err
	}

	return nil
}
	
func %s(ctx context.Context, pool *pgxpool.Pool, cfg *config.Config) error {
	c := containers.NewContainer(cfg)

	if err := c.Container.Invoke(func() {
	// TODO: Implement rollback logic here
	}); err != nil {
		return err
	}

	return nil
}
`, funcNameUp, funcNameDown, funcNameUp, funcNameDown)

	if err := os.WriteFile(seederPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to create seeder file: %w", err)
	}

	fmt.Printf("Successfully created seeder: %s\n", filename)
	return nil
}

func toCamelCase(s string) string {
	if len(s) == 0 {
		return s
	}

	result := ""
	nextUpper := true

	for _, char := range s {
		if char == '_' {
			nextUpper = true
			continue
		}

		if nextUpper {
			if char >= 'a' && char <= 'z' {
				result += string(char - 32)
			} else {
				result += string(char)
			}
			nextUpper = false
		} else {
			result += string(char)
		}
	}

	return toLowerFirst(result)
}

func toLowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	if s[0] >= 'A' && s[0] <= 'Z' {
		return string(s[0]+32) + s[1:]
	}
	return s
}
