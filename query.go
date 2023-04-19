package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func isCmd(item string) bool {
	_, found := cmdMap[item]
	return found
}

const (
	StateOpen = iota
	StateClosed
	StateDeferred
)

type Attrs struct {
	Project      string
	AntiProjects []string
	IDs          []int
	UIDs         []int
	Tags         []string
	AntiTags     []string
	Due          int
	Defer        int
	State        int
	Priority     int
}

type Query struct {
	Cmd        string
	Filters    Attrs
	Attributes Attrs
	Text       string
}

func parseQuery(args ...string) (Query, error) {
	var q = Query{}
	var words []string
	var noMoreIds = false
	var attrs Attrs

	for _, arg := range args {
		item := strings.ToLower(arg)

		if q.Cmd == "" && isCmd(item) {
			q.Cmd = item

			// everything before cmd was filters, after - will be attributes
			q.Filters = attrs
			attrs = Attrs{}
			continue
		}

		if id, err := strconv.ParseInt(item, 10, 64); !noMoreIds && err == nil {
			attrs.IDs = append(attrs.IDs, int(id))
			continue
		}

		if attrs.Project == "" && strings.HasPrefix(item, "+project:") {
			attrs.Project = item[8:]
		} else if strings.HasPrefix(item, "-project:") {
			attrs.AntiProjects = append(attrs.AntiProjects, item[8:])
		} else if strings.HasPrefix(item, "-@") {
			attrs.AntiProjects = append(attrs.AntiProjects, item[2:])
		} else if attrs.Project == "" && strings.HasPrefix(item, "@") {
			attrs.Project = item[1:]
		} else if len(item) > 1 && item[0:1] == "+" {
			attrs.Tags = append(attrs.Tags, item[1:])
		} else if len(item) > 1 && item[0:1] == "-" {
			attrs.AntiTags = append(attrs.AntiTags, item[1:])
		} else if attrs.Priority == 0 && len(item) == 2 && item[0:1] == "p" {
			p, err := strconv.Atoi(arg[1:])
			if err == nil {
				attrs.Priority = p
				continue
			}
		} else if attrs.Due == 0 && len(item) > 4 && strings.HasPrefix(item, "due:") {
			attrs.Due = int(parseDate(resolveAlias(item[4:])))
			fmt.Println(attrs.Due)
		} else if attrs.Defer == 0 && len(item) > 6 && strings.HasPrefix(item, "defer:") {
			attrs.Defer = int(parseDate(resolveAlias(item[6:])))
		} else if len(item) > 6 && strings.HasPrefix(item, "state:") {
			state := item[6:]
			if state == "closed" {
				attrs.State = StateClosed
			}
		} else if len(item) > 4 && strings.HasPrefix(item, "uid:") {
			uid, err := strconv.ParseInt(item[4:], 10, 64)
			if err != nil {
				return q, err
			}
			attrs.UIDs = append(attrs.UIDs, int(uid))
		} else {
			words = append(words, arg)
		}

		noMoreIds = true
	}

	if q.Cmd == "" {
		q.Cmd = "next"
		q.Filters = attrs
	} else {
		q.Attributes = attrs
	}

	q.Text = strings.Join(words, " ")
	return q, nil
}

var dateR = regexp.MustCompile("^(\\d{1,3})([md])$")

func resolveAlias(date string) string {
	if len(date) != 3 {
		return date
	}
	switch date {
	case "tod":
		return "today"
	case "tom":
		return "tomorrow"
	default:
		return date
	}
}

func parseDate(date string) int64 {
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// relative days or months: 40d, 5m, ...
	if len(date) < 5 && dateR.MatchString(date) {
		loc := dateR.FindStringSubmatch(date)
		val, _ := strconv.Atoi(loc[1])
		dur := loc[2]
		if dur == "d" {
			return t.AddDate(0, 0, val).Add(time.Hour*24 - 1).Unix()
		} else if dur == "m" {
			return t.AddDate(0, val, 0).Add(time.Hour*24 - 1).Unix()
		}
	}

	if date == "now" {
		return t.Unix()
	}
	if date == "today" {
		return t.Add(time.Hour*24 - 1).Unix()
	}
	if date == "tomorrow" {
		return t.Add(time.Hour*48 - 1).Unix()
	}
	sw := weekStartDate(t)

	// thisweek, lastweek
	if date == "nextweek" {
		return sw.Add((time.Hour * 24) * 7).Unix()
	}
	// mon | tue | wed | thu | fri | sat | sun
	return 0
}

func weekStartDate(date time.Time) time.Time {
	offset := (int(time.Monday) - int(date.Weekday()) - 7) % 7
	result := date.Add(time.Duration(offset*24) * time.Hour)
	return result
}
