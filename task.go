package main

import (
	"errors"
	"github.com/levmv/tm/table"
	"time"
)

type Task struct {
	Uid       int64
	Id        int64
	Priority  int
	Desc      string
	ProjectId int64
	Created   int64
	Changed   int64
	Started   int64
	Closed    int64
	Deffered  int64
	Time      int64
	Due       int64
}

type TaskView struct {
	Task
	Project         string
	ProjectPriority int
	ProjectUrg      float64
	Tags            string
	TagsCount       int
	Urgency         float64
}

func (t *Task) Start() error {
	if t.isClosed() {
		return errors.New("couldn't start closed task")
	}
	t.Started = time.Now().Unix()
	if err := t.Update(); err != nil {
		return err
	}
	return nil
}

func (t *Task) Stop() error {
	if !t.isActive() {
		return errors.New("task already stopped")
	}
	t.Started = 0
	if err := t.Update(); err != nil {
		return err
	}

	return nil
}

func (t *Task) isActive() bool {
	return t.Started != 0
}

func (t *Task) isClosed() bool {
	return t.Closed != 0
}

func (t *Task) Style() table.Style {
	s := table.Style{
		//Bg: colors.BgDef,
		Fg: colors.FgDef,
	}
	if t.isActive() {
		s.Bg = colors.BgActive
		s.Fg = colors.FgActive
	} else if t.Time > 0 {
		s.Fg = colors.BgPaused
	}

	return s
}

func (t *TaskView) Text() string {
	s := t.Desc
	if t.TagsCount > 0 {
		s = s + " +" + t.Tags
	}
	return s
}

func (t *Task) Create() error {

	t.Created = time.Now().Unix()

	r, err := db.Exec("insert into tasks(`pri`, `desc`, `project_id`, `created`, `due`) values(?,?,?,?,?)",
		t.Priority,
		t.Desc,
		t.ProjectId,
		t.Created,
		t.Due)

	if err != nil {
		return err
	}

	if t.Uid, err = r.LastInsertId(); err != nil {
		panic(err)
	}

	return nil
}

func (t *Task) Update() error {
	_, err := db.Exec("update tasks set `pri`=?,`desc`=?,`project_id`=?,`due`=?,`started`=?,`closed`=? where uid = ?",
		t.Priority,
		t.Desc,
		t.ProjectId,
		t.Due,
		t.Started,
		t.Closed,
		t.Uid)

	if err != nil {
		return err
	}

	return nil
}

func (t *Task) Close() error {
	if _, err := db.Exec("update tasks set id = null, closed = unixepoch('now') where uid = ?", t.Uid); err != nil {
		return err
	}
	return nil
}

func (t *Task) Delete() error {
	if _, err := db.Exec("delete from tasks where uid = ?", t.Uid); err != nil {
		return err
	}
	return nil
}

func (t *TaskView) CalcUrgency() {
	t.Urgency = 3*t.PriorityUrgency() + 2*t.AgeUrgency() + 2*t.ProjectUrgency() + t.TagsUrgency()
}

func (t *TaskView) ProjectUrgency() float64 {
	if t.ProjectId != 0 {
		return t.ProjectUrg
	}
	return 0.1
}

func (t *TaskView) TagsUrgency() float64 {
	if t.TagsCount > 0 {
		return 0.5
	}
	return 0
}

const MaxTaskAge = 60 // TODO: move to config

func (t *Task) AgeUrgency() float64 {
	age := (time.Now().Unix() - t.Created) / (3600 * 24) // age in days
	if age > MaxTaskAge {
		return 1
	}
	return float64(age / MaxTaskAge)
}

func (t *Task) PriorityUrgency() float64 {
	return float64((4 - t.Priority) / 4)
}
