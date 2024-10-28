package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

type period int

const (
	today period = iota
	currentWeek
	month
)

type dateRange struct {
	start, end time.Time
}

func makeDateRange(start, end time.Time) dateRange {
	return dateRange{start: start, end: end}
}

func getCurrentWeek() dateRange {
	t := time.Now()
	midnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)

	offset := (midnight.Weekday() + 7 - 1) % 7
	start := midnight.Add(time.Duration(offset*24) * time.Hour * -1)
	end := start.Add(24*time.Hour*7 - time.Second)
	return makeDateRange(start, end)
}

func getCreatedAt(p period) (string, error) {
	var r dateRange
	switch p {
	case currentWeek:
		r = getCurrentWeek()
	default:
		return "", errors.New("not implemented")
	}
	b := new(bytes.Buffer)
	if err := json.NewEncoder(b).Encode(r); err != nil {
		return "", err
	}
	return b.String(), nil
}
