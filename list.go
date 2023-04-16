package main

import (
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func FindTasks(qf *Attrs) ([]TaskView, error) {
	var (
		r   *sql.Rows
		err error
	)
	r, err = filteredRows(qf)
	if err != nil {
		return nil, err
	}

	var tasks []TaskView
	for r.Next() {
		var t TaskView
		if err := r.Scan(
			&t.Uid,
			&t.Id,
			&t.State,
			&t.Priority,
			&t.Summary,
			&t.ProjectId,
			&t.Project,
			&t.Created,
			&t.Due,
			&t.Started,
			&t.Tags); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	tasks = sortTasks(tasks)

	return tasks, nil
}

func filteredRows(qf *Attrs) (*sql.Rows, error) {
	var conds []string
	if len(qf.IDs) > 0 {
		conds = append(conds, fmt.Sprintf("id IN (%s)", SliceToStr(qf.IDs, ",")))
	}
	if len(qf.UIDs) > 0 {
		conds = append(conds, fmt.Sprintf("uid IN (%s)", SliceToStr(qf.UIDs, ",")))
	}
	if qf.Project != "" {
		p, _ := FindProject(qf.Project)
		conds = append(conds, fmt.Sprintf("`project_id` = %d", p.Id))
	}
	if len(qf.AntiProjects) > 0 {
		ids := ProjectsIds(qf.AntiProjects)
		conds = append(conds, fmt.Sprintf("`project_id` NOT IN %d", ids))
	}
	// todo: tags, antitags

	conds = append(conds, "state = 0")
	conds = append(conds, "deferred < unixepoch('now')")

	return db.Query(fmt.Sprintf(`SELECT uid,id,state,priority,summary,project_id,project,created,due,started,tags 
					 FROM tasks_view WHERE %s ORDER BY created ASC`, strings.Join(conds, " AND ")))
}

func sortTasks(tasks []TaskView) []TaskView {

	for i := range tasks {
		tasks[i].CalcUrgency()
	}

	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Urgency > tasks[j].Urgency
	})
	return tasks
}

func SliceToStr(s []int, sep string) string {
	if len(s) == 0 {
		return ""
	}
	rs := make([]string, len(s))
	for i, v := range s {
		rs[i] = strconv.Itoa(v)
	}
	return strings.Join(rs, sep)
}
