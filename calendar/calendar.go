package calendar

import "time"

type Calendar interface {
	GetEvents()
	PutEvents()
	DeleteAll()
}

type Event struct {
	Title       string
	Start, Stop time.Time
	UID         string
}
