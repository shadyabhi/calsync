package gcal

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"

	googlecalendar "google.golang.org/api/calendar/v3"
)

// Event is the local-representation of googlecalendar.Event
type Event struct {
	*googlecalendar.Event
}

func (e Event) Hash() string {
	var buffer bytes.Buffer
	buffer.WriteString(e.Summary)
	buffer.WriteString(e.Start.DateTime)
	buffer.WriteString(e.End.DateTime)
	buffer.WriteString(e.Description)

	md5sum := md5.Sum(buffer.Bytes())
	return hex.EncodeToString(md5sum[:])
}
