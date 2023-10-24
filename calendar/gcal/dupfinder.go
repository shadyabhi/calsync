package gcal

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
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
	var buffer bytes.Buffer
	buffer.WriteString(gCalEvent.Summary)
	buffer.WriteString(gCalEvent.Start.DateTime)
	buffer.WriteString(gCalEvent.End.DateTime)
	buffer.WriteString(gCalEvent.Description)

	md5sum := md5.Sum(buffer.Bytes())
	eventHash := hex.EncodeToString(md5sum[:])

	_, ok := d.alreadySeen[eventHash]
	if ok {
		log.Printf("Event already processed before, should be a duplicate: %s %s:%s", gCalEvent.Summary, gCalEvent.Start.DateTime, gCalEvent.End.DateTime)
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
