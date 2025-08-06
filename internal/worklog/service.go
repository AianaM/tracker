package worklog

import (
	"fmt"
	"log"
	"time"

	"example.com/tracker/internal/tracker"
	"github.com/AianaM/durationiso8601"
	"github.com/AianaM/timefns"
)

type Rowspan struct {
	Issue   tracker.Issue
	Rowspan int
	Rows    []struct {
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
type Worklogs []tracker.Worklog

func (w Worklogs) asTable(show timefns.TimeSpan) (TableData, error) {
	days := []string{}
	for i := show.Start; i.Before(show.End); i = i.AddDate(0, 0, 1) {
		days = append(days, i.Format(time.DateOnly))
	}
	daysLen := len(days)
	if daysLen == 0 {
		return TableData{}, nil
	}
	daysSums := make([]time.Duration, daysLen)
	rowspans := map[string]Rowspan{}
	sum := time.Duration(0)

	for _, w := range w {
		date, err := timefns.Parse(w.Start)
		if err != nil {
			return TableData{}, fmt.Errorf("error parsing date: %w", err)
		}
		dateStr := date.Format(time.DateOnly)
		for i, v := range days {
			if dateStr != v {
				continue
			}
			if _, ok := rowspans[w.Issue.Key]; !ok {
				rowspans[w.Issue.Key] = Rowspan{w.Issue, 0, []struct {
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
			if date, err := timefns.Parse(w.Start); err != nil {
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

	return TableData{days, rowspans, daysSums, sum}, nil
}

func (h *Handler) getWorklogsTable(createdBy string, timespan, show titledTimeSpan) (TableData, error) {
	worklogs, err := h.trackerClient.GetWorklog(createdBy, timespan.timespan)
	if err != nil {
		return TableData{}, fmt.Errorf("error getting worklogs: %w", err)
	}
	return Worklogs(worklogs).asTable(show.timespan)
}
