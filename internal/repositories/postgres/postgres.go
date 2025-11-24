package postges

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/nikomkinds/SchoolSchedule/internal/config"
)

// NewPostgresDB connects to a database using params loaded previously in config
func NewPostgresDB(cfg *config.Config) (*sql.DB, error) {

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("unalble to open database connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("unable to ping database^ %w", err)
	}

	return db, nil
}
