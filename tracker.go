package main

import "time"

type Tracker interface {
	getWorklog(from time.Time, to time.Time) ([]worklog, error)
}
