package main

import (
	"time"
)

// информация о текущем пользователе
type Myself struct {
	self                                                                 string
	uid                                                                  int
	login                                                                string
	trackerUid, passportUid                                              int
	cloudUid, firstName, lastName, display, email                        string
	external, hasLicense, dismissed, useNewFilters, disableNotifications bool
	firstLoginDate, lastLoginDate                                        time.Time
	welcomeMailSent                                                      bool
}
