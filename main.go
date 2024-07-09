package main

import (
	"database/sql"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbFile := getEnv("TODO_DBFILE", "./scheduler.db")
	install := !isExistDb(dbFile)

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		log.Fatalf("Ошибка при открытии базы данных: %v", err)
	}
	defer db.Close()

	if install {
		if err := installDb(db); err != nil {
			log.Fatalf("Ошибка при установке базы данных: %v", err)
		}
	}

	store := NewStore(db)
	service := NewService(store)
	handler := NewHandler(service)

	server := Server{}
	server.Run(getEnv("TODO_PORT", "7540"), handler.InitRouter())
}

func isExistDb(dbPath string) bool {
	_, err := os.Stat(dbPath)
	return err == nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func installDb(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL,
		title TEXT NOT NULL,
		comment TEXT,
		repeat TEXT
	);
	CREATE INDEX idx_date ON scheduler(date);
	`
	_, err := db.Exec(createTableSQL)
	if err != nil {
		return err
	}
	return nil
}
