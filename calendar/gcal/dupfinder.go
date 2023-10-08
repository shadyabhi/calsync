package gcal

import (
	"log"

	"calsync/calendar"

	googlecalendar "google.golang.org/api/calendar/v3"
)

type DuplicateEventsFinder struct {
	alreadySeen map[string]bool
}

func newDuplicateEventsFinder() *DuplicateEventsFinder {
	return &DuplicateEventsFinder{
		alreadySeen: make(map[string]bool),
	}
}

func (d *DuplicateEventsFinder) isGCalinEvents(gCalEvent *googlecalendar.Event, events []calendar.Event) (bool, int) {
	eventHash := gCalEvent.Summary + gCalEvent.Start.DateTime + gCalEvent.End.DateTime
	_, ok := d.alreadySeen[eventHash]
	if ok {
		log.Printf("Event already processed before, should be a duplicate: %s %s:%s", gCalEvent.Summary, gCalEvent.Start.DateTime, gCalEvent.End.DateTime)
		return false, -1
	}

	d.alreadySeen[eventHash] = true
	for i, e := range events {
		if e.EventHash() == eventHash {
			return true, i
		}
	}

	return false, -1
}
