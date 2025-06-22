package main

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/AianaM/durationiso8601"
	"github.com/AianaM/timefns"
)

// записи о затраченном времени
type worklog struct {
	Self      string `json:"self"`
	ID        int    `json:"id"`
	Version   int    `json:"version"`
	Issue     issue  `json:"issue"`
	Comment   string `json:"comment"`
	CreatedBy user   `json:"createdBy"`
	UpdatedBy user   `json:"updatedBy"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Start     string `json:"start"`
	Duration  string `json:"duration"`
}

type issue struct {
	Self    string `json:"self"`
	Id      string `json:"id"`
	Key     string `json:"key"`
	Display string `json:"display"`
}

type user struct {
	Self    string `json:"self"`
	Id      string `json:"id"`
	Display string `json:"display"`
}
type Rowspan struct {
	IssueKey string
	Rowspan  int
	Rows     []struct {
		Comment  string
		Duration []string
	}
	Sum time.Duration
}
type TableData struct {
	Days     []string
	Rowspans map[string]Rowspan
	DaysSum  []time.Duration
	Sum      time.Duration
}

func (c cloud) getWorklog(ts timefns.TimeSpan) (TableData, error) {
	log.Println("getWeekWorklog", ts.Start, ts.End)
	request := newRequestData(
		"https://api.tracker.yandex.net/v2/worklog",
		map[string]string{"X-Cloud-Org-Id": c.orgId},
		[]keyValue{
			{"createdBy", strconv.FormatInt(int64(c.id), 10)},
			{"createdAt", "from:" + ts.Start.Format(time.RFC3339Nano)},
			{"createdAt", "to:" + ts.End.Format(time.RFC3339Nano)},
		},
		[]worklog{},
	)
	if err := request.get(); err != nil {
		log.Fatal(err)
		return TableData{}, err
	}

	return getWorklogsTable(ts, request.body), nil
}

func getWorklogsTable(ts timefns.TimeSpan, worklogs []worklog) TableData {
	days := []string{}
	for i := ts.Start; i.Before(ts.End); i = i.AddDate(0, 0, 1) {
		days = append(days, i.Format(time.DateOnly))
	}
	daysLen := len(days)
	if daysLen == 0 {
		return TableData{}
	}
	daysSums := make([]time.Duration, daysLen)
	rowspans := map[string]Rowspan{}
	sum := time.Duration(0)

	for _, w := range worklogs {
		date, err := timefns.Parse(w.Start)
		if err != nil {
			log.Fatal(err)
		}
		dateStr := date.Format(time.DateOnly)
		for i, v := range days {
			if dateStr != v {
				continue
			}
			if _, ok := rowspans[w.Issue.Key]; !ok {
				rowspans[w.Issue.Key] = Rowspan{w.Issue.Key, 0, []struct {
					Comment  string
					Duration []string
				}{}, time.Duration(0)}
			}
			newRow := make([]string, daysLen)
			newRow[i] = w.Duration
			rowspan := rowspans[w.Issue.Key]
			rowspan.Rowspan++
			rowspan.Rows = append(rowspan.Rows, struct {
				Comment  string
				Duration []string
			}{w.Comment, newRow})
			if date, err := timefns.Parse(w.CreatedAt); err != nil {
				log.Println("Error parsing date:", err)
			} else if duration, err := durationiso8601.ParseDuration(date, w.Duration); err != nil {
				log.Println("Error parsing duration:", err)
			} else {
				rowspan.Sum += duration
				daysSums[i] += duration
				sum += duration
			}

			rowspans[w.Issue.Key] = rowspan
			break
		}
	}

	return TableData{days, rowspans, daysSums, sum}
}

func DurationBeautify(d time.Duration) string {
	h := int(d / time.Hour)
	m := int((d % time.Hour) / time.Minute)

	parts := make([]string, 0, 2)
	if h > 0 {
		parts = append(parts, strconv.Itoa(h)+"h")
	}

	if m > 0 {
		parts = append(parts, strconv.Itoa(m)+"m")
	}

	if len(parts) == 0 {
		return "0m"
	}

	return strings.Join(parts, " ")
}
