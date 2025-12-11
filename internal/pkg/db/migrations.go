package db

import (
	"database/sql"
)

// EnsureAuthTables creates minimal auth-related tables that the service expects.
// This is intentionally narrow: it only creates the `refresh_tokens` table if missing
// so running the service will not fail when the SQL init scripts were not applied.
func EnsureAuthTables(db *sql.DB) error {
	const q = `CREATE TABLE IF NOT EXISTS refresh_tokens (
  id BIGINT AUTO_INCREMENT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  token_hash VARCHAR(128) NOT NULL,
  expires_at DATETIME NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX (token_hash),
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);`
	_, err := db.Exec(q)
	return err
}
