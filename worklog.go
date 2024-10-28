package main

import (
	"fmt"
	"log"
)

// записи о затраченном времени
type worklog struct {
	self                                  string
	id                                    int
	version                               int
	issue                                 issue
	comment                               string
	createdBy, updatedBy                  user
	createdAt, updatedAt, start, duration string
}

type issue struct {
	self, id, key, display string
}

type user struct {
	self, id, display string
}

func (c cloud) getWeekWorklog() {
	p, err := getCreatedAt(currentWeek)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println(p)
	w := makeRequestData(
		"https://api.tracker.yandex.net/v2/worklog",
		map[string]string{"X-Cloud-Org-Id": c.orgId},
		map[string]string{"createdBy": c.id, "createdAt": p},
		[]worklog{},
	)
	if err := w.get(); err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println(w.body)
}
