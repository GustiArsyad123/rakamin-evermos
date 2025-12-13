package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func NewMySQL() (*sql.DB, error) {
	// Load local `.env` file if present (do not override existing env values).
	// This is a lightweight loader to avoid adding external dependencies.
	if _, err := os.Stat(".env"); err == nil {
		if b, err := os.ReadFile(".env"); err == nil {
			for _, line := range strings.Split(string(b), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}
				k := strings.TrimSpace(parts[0])
				v := strings.Trim(strings.TrimSpace(parts[1]), "'\"")
				if os.Getenv(k) == "" {
					_ = os.Setenv(k, v)
				}
			}
		}
	}

	user := getenv("DB_USER", "user")
	// Prefer DB_PASS env var. If empty, support reading password from a file
	// path provided in `DB_PASS_FILE` (useful for docker secrets).
	pass := os.Getenv("DB_PASS")
	if pass == "" {
		if f := os.Getenv("DB_PASS_FILE"); f != "" {
			if b, err := os.ReadFile(f); err == nil {
				pass = strings.TrimSpace(string(b))
			}
		}
	}
	// fallback default if nothing provided
	if pass == "" {
		pass = getenv("DB_PASS", "password")
	}

	host := getenv("DB_HOST", "127.0.0.1")
	port := getenv("DB_PORT", "3307")
	name := getenv("DB_NAME", "ecommerce")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, name)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pooling for optimal performance
	db.SetMaxOpenConns(25)                 // Maximum open connections
	db.SetMaxIdleConns(10)                 // Maximum idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum connection lifetime

	if err := db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func getenv(k, fallback string) string {
	v := os.Getenv(k)
	if v == "" {
		return fallback
	}
	return v
}
