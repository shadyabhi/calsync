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

func (e Event) String() string {
	return e.Title + " " + e.Start.Format(time.RFC3339) + " " + e.Stop.Format(time.RFC3339) + " " + e.UID
}
