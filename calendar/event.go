package calendar

import "time"

// Event describes an event in calendar
type Event struct {
	Title       string
	Start, Stop time.Time
	UID         string
}

// Calendar interface should be satisfied for a calendar provider
// Ex. Google Calendar
