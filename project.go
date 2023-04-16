package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type Project struct {
	Id        int64
	Name      string
	Priority  int64
	Open      int64
	Closed    int64
	TimeTotal int64
}

func (p *Project) Save() error {
	r, err := db.Exec("INSERT INTO projects(`name`) VALUES(?)", p.Name)
	if err != nil {
		return err
	}

	if p.Id, err = r.LastInsertId(); err != nil {
		panic(err)
	}

	return nil
}

func FindProject(name string) (*Project, error) {
	p := Project{}
	row := db.QueryRow("SELECT id, name FROM projects WHERE name = ?", name)
	if err := row.Scan(&p.Id, &p.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// ProjectID returns project id by name. Creates project if not exist
func ProjectID(name string) int64 {
	var id int64
	row := db.QueryRow("SELECT id FROM projects WHERE name = ?", name)
	err := row.Scan(&id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			panic(err)
		}
		id = newProject(name)
	}
	return id
}

func Projects() ([]Project, error) {
	r, err := db.Query(fmt.Sprintf("SELECT * FROM projects"))
	if err != nil {
		return nil, err
	}
	var ps []Project
	for r.Next() {
		var p Project
		if err := r.Scan(
			&p.Id, &p.Name, &p.Priority, &p.Open, &p.Closed, &p.TimeTotal); err != nil {
			panic(err)
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func ProjectsIds(names []string) []int64 {
	r, err := db.Query(fmt.Sprintf("SELECT id FROM projects WHERE name IN (%s)", strings.Join(names, ",")))
	if err != nil {
		panic(err)
	}
	var ids []int64
	for r.Next() {
		var id int64
		if err := r.Scan(&id); err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}
	return ids
}

func newProject(name string) int64 {
	var id int64

	r, err := db.Exec("INSERT INTO projects(`name`) VALUES(?)", name)
	if err != nil {
		panic(err)
	}

	id, err = r.LastInsertId()
	if err != nil {
		panic(err)
	}

	return id
}
