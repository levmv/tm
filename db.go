package main

import (
	"database/sql"
	"fmt"
	"os"
)
import _ "modernc.org/sqlite"

func initDb() (*sql.DB, error) {

	path, err := getFilePath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)

	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		return nil, err
	}

	return db, nil
}

func getFilePath() (string, error) {
	path := os.Getenv("TM_DB")

	if path != "" {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return home + "/.tm.sqlite", nil
}
