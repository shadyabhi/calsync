package gcal

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"time"

	googlecalendar "google.golang.org/api/calendar/v3"
)

// Event is the local-representation of googlecalendar.Event
type Event struct {
	*googlecalendar.Event
}

func (e Event) Hash() string {
	var buffer bytes.Buffer
	buffer.WriteString(e.Summary)

	startTime, _ := time.Parse(time.RFC3339, e.Start.DateTime)
	endTime, _ := time.Parse(time.RFC3339, e.End.DateTime)

	buffer.WriteString(startTime.UTC().Format(time.RFC3339))
	buffer.WriteString(endTime.UTC().Format(time.RFC3339))
	buffer.WriteString(e.Description)

	md5sum := md5.Sum(buffer.Bytes())
	return hex.EncodeToString(md5sum[:])
}
