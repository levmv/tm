package main

import (
	"database/sql"
	"errors"
)

type Tag struct {
	ID   int64
	Name string
}

// TagID returns tag id by name. Create tag if not exist
func TagID(name string) int64 {
	var id int64
	row := db.QueryRow("SELECT id FROM tags WHERE name = ?", name)
	err := row.Scan(&id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			panic(err)
		}
		id = newTag(name)
	}
	return id
}

func LinkTag(name string, taskId int64) error {
	tagId := TagID(name)
	_, err := db.Exec("INSERT INTO tags_tasks(tag_id, task_id) VALUES(?, ?)", tagId, taskId)
	if err != nil {
		return err
	}
	return nil
}

func newTag(name string) int64 {
	var id int64

	r, err := db.Exec("INSERT INTO tags(`name`) VALUES(?)", name)
	if err != nil {
		panic(err)
	}

	id, err = r.LastInsertId()
	if err != nil {
		panic(err)
	}

	return id
}
