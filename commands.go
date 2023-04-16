package main

import (
	"errors"
	"fmt"
)

type CommandFunc func(q *Query) error

func CommandAdd(q *Query) error {
	task := Task{}
	task.Summary = q.Text

	if q.Attributes.Project != "" {
		task.ProjectId = ProjectID(q.Attributes.Project)
	}
	if q.Attributes.Priority != 0 {
		task.Priority = q.Attributes.Priority
	}
	if q.Attributes.Due != 0 {
		task.Due = int64(q.Attributes.Due)
	}

	if err := task.Create(); err != nil {
		return err
	}

	if len(q.Attributes.Tags) > 0 {
		for _, tag := range q.Attributes.Tags {
			LinkTag(tag, task.Uid)
		}
	}
	fmt.Printf("new task %v created\n", task.Uid)
	return nil
}

func CommandModify(q *Query) error {
	tasks, err := FindTasks(&q.Filters)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if q.Attributes.Project != "" {
			t.ProjectId = ProjectID(q.Attributes.Project)
		}
		if q.Attributes.Priority != 0 {
			t.Priority = q.Attributes.Priority
		}

		if err := t.Update(); err != nil {
			return err
		}

		if len(q.Attributes.Tags) > 0 {
			for _, tag := range q.Attributes.Tags {
				LinkTag(tag, t.Uid)
			}
		}
	}
	return nil
}

func CommandNext(q *Query) error {
	tasks, err := FindTasks(&q.Filters)
	if err != nil {
		return err
	}

	if err := Display(tasks); err != nil {
		return err
	}
	return nil
}

func CommandStart(q *Query) error {
	tasks, err := FindTasks(&q.Filters)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := t.Start(); err != nil {
			return err
		}
		fmt.Printf("Task %v started\n", t.Id)
	}
	return nil
}

func CommandStop(q *Query) error {
	tasks, err := FindTasks(&q.Filters)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := t.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func CommandRemove(q *Query) error {
	tasks, err := FindTasks(&q.Filters)
	// todo: confirm if more then one
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := t.Delete(); err != nil {
			fmt.Printf("#%v (%v) deleted\n", t.Id, t.Summary)
			return err
		}
	}
	return nil
}

func CommandDone(q *Query) error {
	tasks, err := FindTasks(&q.Filters)
	if err != nil {
		return err
	}
	for _, t := range tasks {
		if err := t.Done(); err != nil {
			fmt.Printf("#%v (%v) resolved\n", t.Id, t.Summary)
			return err
		}
	}
	return nil
}

func CommandCount(q *Query) error {
	tasks, err := FindTasks(&q.Filters)
	if err != nil {
		return err
	}
	fmt.Printf("%d tasks found\n", len(tasks))

	return nil
}

func CommandInfo(q *Query) error {
	if len(q.Filters.IDs) != 1 {
		return errors.New("must be selected one task")
	}
	tasks, err := FindTasks(&q.Filters)
	if err != nil {
		return err
	}
	if err := DisplayOne(tasks[0]); err != nil {
		Error("%v", err)
	}
	return nil
}

func CommandProjects(q *Query) error {
	list, _ := Projects()

	if err := DisplayProjects(list); err != nil {
		return err
	}
	return nil
}
