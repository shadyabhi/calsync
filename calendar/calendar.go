package calendar

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"sort"
	"time"
)

type Calendar interface {
	// String returns a string representation of the calendar.
	// This is used for logging and debugging purposes.
	String() string

	// GetEvents retrieves events from the calendar within the specified time range.
	GetEvents(start time.Time, end time.Time) ([]Event, error)

	// PutEvents adds or updates events in the calendar.
	PutEvents() error

	// DeleteAll removes all events from the calendar that were created by calsync.
	// The nDays parameter specifies how many days back to look for events to delete.
	DeleteAll(nDays int) error

	// SyncToDest synchronizes events to the destination calendar.
	SyncToDest([]Event) error
}

type Event struct {
	Title       string
	Notes       string
	Start, Stop time.Time
	UID         string
}

type Events []Event

// SortStartTime events by start time
func (e Events) SortStartTime() {
	sort.Slice(e, func(i, j int) bool {
		return e[i].Start.Before(e[j].Start)
	})
}

func (e Event) String() string {
	return e.Title + " " + e.Start.UTC().String() + " " + e.Stop.UTC().String() + " " + e.UID
}

func (e Event) Hash() string {
	var buffer bytes.Buffer
	buffer.WriteString(e.Title)
	buffer.WriteString(e.Start.UTC().Format(time.RFC3339))
	buffer.WriteString(e.Stop.UTC().Format(time.RFC3339))
	buffer.WriteString(e.Notes)

	md5sum := md5.Sum(buffer.Bytes())

	return hex.EncodeToString(md5sum[:])
}
