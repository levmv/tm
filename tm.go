package main

import (
	"database/sql"
	"fmt"
	"os"
)

const (
	Version = "0.0.1"
)

var (
	db *sql.DB

	cmdMap = map[string]CommandFunc{
		"next":     CommandNext,
		"add":      CommandAdd,
		"modify":   CommandModify,
		"remove":   CommandRemove,
		"start":    CommandStart,
		"stop":     CommandStop,
		"done":     CommandDone,
		"count":    CommandCount,
		"info":     CommandInfo,
		"projects": CommandProjects,
	}
)

func run(q Query) int {
	var err error
	db, err = initDb()
	if err != nil {
		return Error("initDB failed: %w", err)
	}
	defer db.Close()

	if err := ApplyMigrations(); err != nil {
		return Error("failed to apply migrations: %w", err)
	}

	cmd := cmdMap[q.Cmd]
	if err := cmd(&q); err != nil {
		return Error("error: %w", err)
	}

	return 0
}

func main() {

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "help":
			fmt.Println("help")
			return
		case "version":
			fmt.Println(Version)
			return
		}
	}

	q, err := parseQuery(os.Args[1:]...)
	if err != nil {
		Error("error: %w", err)
		os.Exit(1)
	}

	os.Exit(run(q))
}
