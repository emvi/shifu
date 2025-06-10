package db

import (
	"github.com/emvi/shifu/pkg/admin/util"
	"github.com/emvi/shifu/pkg/cfg"
	"log/slog"
)

var migrations = []string{
	`CREATE TABLE "migrations" (version int NOT NULL PRIMARY KEY);

CREATE TABLE "user" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email VARCHAR(320) NOT NULL UNIQUE,
    password VARCHAR(60) NOT NULL,
	password_salt VARCHAR(20) NOT NULL,
	full_name VARCHAR(100) NOT NULL
);

CREATE TABLE "session" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id int NOT NULL REFERENCES "user" (id),
	session VARCHAR(40) NOT NULL UNIQUE,
	expires timestamp NOT NULL
);

CREATE TABLE "reference" (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	name VARCHAR(500) NOT NULL UNIQUE
);`,
}

func migrate() {
	var exists bool

	if err := db.Get(&exists, "SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE name = 'migrations')"); err != nil {
		panic(err)
	}

	latestVersion := -1

	if exists {
		if err := db.Get(&latestVersion, "SELECT max(version) FROM migrations"); err != nil {
			panic(err)
		}

		slog.Info("Database found", "version", latestVersion)
	} else {
		slog.Info("Setting up new database")
	}

	for i := latestVersion + 1; i < len(migrations); i++ {
		if _, err := db.Exec(migrations[i]); err != nil {
			panic(err)
		}

		if i == 0 {
			salt := util.GenRandomString(20)
			password := ""

			if cfg.Get().UI.AdminPassword != "" {
				password = cfg.Get().UI.AdminPassword
			} else {
				password = util.GenRandomString(10)
				slog.Info("Creating admin user", "password", password)
			}

			pwd := util.HashPassword(password + salt)

			if _, err := db.Exec(`INSERT INTO "user" (email, password, password_salt, full_name) VALUES ('admin', ?, ?, 'Admin')`, pwd, salt); err != nil {
				panic(err)
			}
		}

		if _, err := db.Exec("INSERT INTO migrations (version) VALUES (?)", i); err != nil {
			panic(err)
		}
	}
}
