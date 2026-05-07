package config

import (
	"os"
	"path/filepath"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sqlx.DB

func InitDB() error {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/example.db"
	}

	err := os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to create data directory")
		return err
	}

	db, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to connect to database")
		return err
	}

	err = db.Ping()
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to ping database")
		return err
	}

	err = enableWAL(db)
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to enable WAL mode")
		return err
	}

	DB = db
	err = createTables()
	if err != nil {
		Logger.Error().Err(err).Msg("Failed to create tables")
		return err
	}

	Logger.Info().Msg("Database initialized successfully")
	return nil
}

func enableWAL(db *sqlx.DB) error {
	_, err := db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA cache_size=-10000;")
	if err != nil {
		return err
	}

	Logger.Info().Msg("WAL mode enabled successfully")
	return nil
}

func createTables() error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		nickname TEXT,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		birthday DATE,
		sign TEXT,
		status INTEGER DEFAULT 1,
		create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		update_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS chat_rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		logo TEXT,
		"desc" TEXT,
		owner_id INTEGER,
		"group" TEXT,
		status INTEGER DEFAULT 1,
		create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		update_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (owner_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_id INTEGER,
		sender INTEGER,
		notify TEXT,
		message TEXT NOT NULL,
		send_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (room_id) REFERENCES chat_rooms(id),
		FOREIGN KEY (sender) REFERENCES users(id)
	);
	`

	_, err := DB.Exec(schema)
	if err != nil {
		return err
	}

	return nil
}
