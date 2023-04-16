package main

import (
	"fmt"
	"github.com/levmv/tm/table"
	"os"
	"strconv"
	"time"
)

type ColorScheme struct {
	FgDefault int
	BgDefault int
	FgActive  int
	BgActive  int
	BgPaused  int
}

var colors = ColorScheme{
	FgDefault: 250,
	BgDefault: 233,
	FgActive:  233,
	BgActive:  250,
	BgPaused:  245,
}

func Warning(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "\u001b[33m"+format+"\u001B[0m\n", a...)
}

func Error(format string, a ...any) int {
	_, _ = fmt.Fprintf(os.Stderr, "\u001b[31m"+format+"\u001B[0m\n", a...)
	return 1
}

func Display(tasks []TaskView) error {

	tbl := table.New([]string{
		"ID",
		"Project",
		"Age",
		"Due",
		"U",
		"Summary",
	}, table.Style{
		Fg: colors.FgDefault,
		Bg: colors.BgDefault,
	})

	now := time.Now().Unix()

	for _, t := range tasks {
		tbl.AddRow(table.Row{
			Cells: []string{
				strconv.Itoa(int(t.Id)),
				t.Project,
				ageString(int(now - t.Created)),
				niceDate(t.Due),
				fmt.Sprintf("%.2f", t.Urgency),
				t.Text(),
			},
			Style: t.Style(),
		})
	}

	tbl.Render()

	return nil
}

func DisplayProjects(list []Project) error {
	tbl := table.New([]string{
		"Name",
		"Tasks",
		"Time",
	}, table.Style{
		Fg: colors.FgDefault,
		Bg: colors.BgDefault,
	})

	for _, t := range list {
		tbl.AddRow(table.Row{
			Cells: []string{
				t.Name,
				fmt.Sprintf("%d/%d", t.Open, t.Closed),
				fmt.Sprintf("%d", t.TimeTotal),
			},
			Style: table.Style{},
		})
	}

	tbl.Render()
	return nil
}

func DisplayOne(task TaskView) error {

	fmt.Printf("%10s %v\n", "ID", task.Id)
	fmt.Printf("%10s %v\n", "UID", task.Uid)
	fmt.Printf("%10s %v\n", "Project", task.Project)
	fmt.Printf("%10s %v\n", "Summary", task.Summary)
	fmt.Printf("%10s %v\n", "Created", time.Unix(task.Created, 0).Format(time.RFC850))
	fmt.Printf("%10s %v\n", "Updated", time.Unix(task.Created, 0).Format(time.RFC850))
	if task.Started > 0 {
		fmt.Printf("%10s %v ago\n", "Started", ageString(int(task.Started)))
	}
	if task.Due > 0 {
		fmt.Printf("%10s %v\n", "Due", time.Unix(task.Due, 0).Format(time.RFC850))
	}

	return nil
}

func ageString(age int) string {
	if age < 65 {
		return fmt.Sprintf("%ds", age)
	}
	if age < 3600 {
		return fmt.Sprintf("%dm", age/60)
	}
	if age < 3600*24*2 {
		return fmt.Sprintf("%dh", age/3600)
	}
	if age < 3600*24*60 {
		return fmt.Sprintf("%dd", age/(3600*24))
	}
	return fmt.Sprintf("%dm", age/(3600*24*30))
}

func niceDate(d int64) string {
	if d == 0 {
		return ""
	}
	return time.Unix(d, 0).Format("Jan 06")
}
