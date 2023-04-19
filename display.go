package main

import (
	"bufio"
	"fmt"
	"github.com/levmv/tm/table"
	"os"
	"strconv"
	"strings"
	"time"
)

type ColorScheme struct {
	FgDef    int
	BgDef    int
	AltBgDef int
	FgActive int
	BgActive int
	BgPaused int
}

var colors = ColorScheme{
	FgDef:    250,
	BgDef:    232,
	AltBgDef: 233,
	FgActive: 233,
	BgActive: 250,
	BgPaused: 245,
}

func Warning(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "\u001b[33mwarning: "+format+"\u001B[0m\n", a...)
}

func Error(format string, a ...any) int {
	_, _ = fmt.Fprintf(os.Stderr, "\u001b[31merror: "+format+"\u001B[0m\n", a...)
	return 1
}

func Confirm(format string, a ...any) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(format+" [y/n]:\n", a...)

		r, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}

		r = strings.ToLower(strings.TrimSpace(r))
		if r == "y" || r == "yes" {
			return true
		} else if r == "n" || r == "no" {
			return false
		}
	}
}

func DisplayTasks(tasks []TaskView) error {

	tbl := table.New([]string{
		"ID",
		"Project",
		"Age",
		"Due",
		"U",
		"Summary",
	}, table.Style{
		Fg:    colors.FgDef,
		Bg:    colors.BgDef,
		AltBg: colors.AltBgDef,
	})

	tbl.SetColStyle(1, table.Style{
		Fg: 6,
		Bg: 0,
	})

	now := time.Now().Unix()

	for _, t := range tasks {
		tbl.AddRow([]string{
			strconv.Itoa(int(t.Id)),
			t.Project, //fmt.Sprintf("\u001b[33m%v\u001B[0m", t.Project),
			ageString(int(now - t.Created)),
			niceDate(t.Due),
			fmt.Sprintf("%.2f", t.Urgency),
			t.Text(),
		}, t.Style())
	}

	tbl.Render()

	return nil
}

func DisplayProjects(list []Project) error {
	tbl := table.New([]string{
		"Name",
		"Tasks",
		"Time",
		"Rank",
	}, table.Style{
		Fg: colors.FgDef,
		Bg: colors.BgDef,
	})

	for _, t := range list {
		tbl.AddRow([]string{
			t.Name,
			fmt.Sprintf("%d/%d", t.Open, t.Closed),
			fmt.Sprintf("%d", t.Time),
			fmt.Sprintf("%.2f", t.Urgency),
		},
			table.Style{},
		)
	}

	tbl.Render()
	return nil
}

func DisplayOne(task TaskView) error {

	fmt.Printf("%10s %v\n", "ID", task.Id)
	fmt.Printf("%10s %v\n", "UID", task.Uid)
	fmt.Printf("%10s %v\n", "Project", task.Project)
	fmt.Printf("%10s %v\n", "Desc", task.Desc)
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
