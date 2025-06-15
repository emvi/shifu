package db

import (
	"github.com/emvi/shifu/pkg/cfg"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"path/filepath"
)

var db *sqlx.DB

// Connect connects to SQLite.
func Connect() {
	path := filepath.Join(cfg.Get().BaseDir, "shifu.db")
	slog.Info("Connecting to database", "path", path)
	var err error
	db, err = sqlx.Open("sqlite3", path)

	if err != nil {
		slog.Error("Error connecting to admin database", "error", err)
		panic(err)
	}

	migrate()
}

// Disconnect closes the SQLite connection.
func Disconnect() {
	slog.Info("Disconnecting from database")

	if err := db.Close(); err != nil {
		slog.Error("Error closing admin database", "error", err)
	}
}

// Get returns the database connection.
func Get() *sqlx.DB {
	return db
}
