package main

import (
	_ "embed"
	"fmt"
)

type Migration func() error

var migrations = []Migration{
	Base,
}

func getDbVer() (int, error) {
	var id int
	r := db.QueryRow("PRAGMA user_version")
	err := r.Scan(&id)
	return id, err
}

func setDbVer(version int) error {
	_, err := db.Query(fmt.Sprintf("PRAGMA user_version = %d", version))
	return err
}

func ApplyMigrations() error {
	curVer, err := getDbVer()
	if err != nil {
		return err
	}
	for i, m := range migrations {
		if i < curVer {
			continue
		}
		tx, _ := db.Begin()
		err := m()
		if err != nil {
			tx.Rollback()
			return err
		} else {
			tx.Commit()
		}
		if err := setDbVer(i + 1); err != nil {
			return err
		}

	}
	return nil
}

//go:embed sql/base.sql
var baseSql string

func Base() error {
	if _, err := db.Query(baseSql); err != nil {
		return err
	}
	return nil
}
