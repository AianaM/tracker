package main

import (
	"fmt"
	"time"
)

type Tracker interface {
	getWeekWorklog()
}

var c cloud

func init() {
	c = makeClouds()

	if loc, err := time.LoadLocation("Europe/Moscow"); err != nil {
		fmt.Println(err)
	} else {
		time.Local = loc
	}
}

func main() {
	c.getWeekWorklog()
}
