package main

import (
	"errors"
	"github.com/levmv/tm/table"
	"time"
)

const (
	StateOpen = iota
	StateClosed
	StateDeferred
)

type Task struct {
	Uid       int64
	Id        int64
	State     int
	Priority  int
	Summary   string
	ProjectId int64
	Created   int64
	Changed   int64
	Started   int64
	Closed    int64
	Deffered  int64
	TimeSpent int64
	Due       int64
}

type TaskView struct {
	Task
	Project         string
	ProjectPriority int
	Tags            string
	TagsCount       int
	Urgency         float64
}

func (t *Task) Start() error {
	if t.State != StateOpen {
		return errors.New("couldn't start closed task")
	}
	t.Started = time.Now().Unix()
	if err := t.Update(); err != nil {
		return err
	}
	return nil
}

func (t *Task) Stop() error {
	if t.Started == 0 {
		return errors.New("task already stopped")
	}
	t.Started = 0
	if err := t.Update(); err != nil {
		return err
	}

	return nil
}

func (t *Task) Done() error {
	t.State = StateClosed
	if err := t.Update(); err != nil {
		return err
	}
	return nil
}

func (t *Task) Style() table.Style {
	s := table.Style{
		Bg: colors.BgDefault,
		Fg: colors.FgDefault,
	}
	if t.Started > 0 {
		s.Bg = colors.BgActive
		s.Fg = colors.FgActive
	} else if t.TimeSpent > 0 {
		s.Fg = colors.BgPaused
	}

	return s
}

func (t *TaskView) Text() string {
	s := t.Summary
	if t.TagsCount > 0 {
		s = s + " +" + t.Tags
	}
	return s
}

func (t *Task) Create() error {

	t.Created = time.Now().Unix()

	r, err := db.Exec("INSERT INTO tasks(`state`, `priority`, `summary`, `project_id`, `created`, `due`) VALUES(?,?,?,?,?,?)",
		t.State,
		t.Priority,
		t.Summary,
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
	_, err := db.Exec("UPDATE tasks SET `state`=?, `priority`=?,`summary`=?,`project_id`=?,`due`=?,`started`=? WHERE uid = ?",
		t.State,
		t.Priority,
		t.Summary,
		t.ProjectId,
		t.Due,
		t.Started,
		t.Uid)

	if err != nil {
		return err
	}

	return nil
}

func (t *Task) Delete() error {
	if _, err := db.Exec("DELETE FROM tasks WHERE uid = ?", t.Uid); err != nil {
		return err
	}
	return nil
}

func (t *TaskView) CalcUrgency() {
	t.Urgency = 3*t.PriorityUrgency() + 2*t.AgeUrgency() + t.ProjectUrgency() + t.TagsUrgency()
}

func (t *TaskView) ProjectUrgency() float64 {
	if t.ProjectId != 0 {
		return float64((4 - t.ProjectPriority) / 4)
	}
	return 0.25
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
