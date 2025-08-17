package calendar

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"sort"
	"time"
)

type Calendar interface {
	GetEvents(start time.Time, end time.Time) ([]Event, error)
	PutEvents() error
	DeleteAll(nDays int) error

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
