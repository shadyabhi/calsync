package gcal

import (
	"log"

	"calsync/calendar"
)

type DuplicateEventsFinder struct {
	alreadySeen map[string]bool
}

func newDuplicateEventsFinder() *DuplicateEventsFinder {
	return &DuplicateEventsFinder{
		alreadySeen: make(map[string]bool),
	}
}

func (d *DuplicateEventsFinder) isGCalinEvents(event *Event, events []calendar.Event) (bool, int) {
	eventHash := event.Hash()

	_, ok := d.alreadySeen[eventHash]
	if ok {
		log.Printf("Event already processed before, should be a duplicate: %s %s:%s", event.Summary, event.Start.DateTime, event.End.DateTime)
		return false, -1
	}

	d.alreadySeen[eventHash] = true
	for i, e := range events {
		if e.Hash() == eventHash {
			return true, i
		}
	}

	return false, -1
}
