package calendar

import "time"

type Calendar interface {
	GetEvents()
	PutEvents()
}

type Event struct {
	Title       string
	Start, Stop time.Time
	UID         string
}
