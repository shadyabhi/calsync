package calendar

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"time"
)

type Calendar interface {
	GetEvents()
	PutEvents()
	DeleteAll()
}

type Event struct {
	Title       string
	Notes       string
	Start, Stop time.Time
	UID         string
}

func (e Event) String() string {
	return e.Title + " " + e.Start.UTC().String() + " " + e.Stop.UTC().String() + " " + e.UID
}

func (e Event) Hash() string {
	var buffer bytes.Buffer
	buffer.WriteString(e.Title)
	buffer.WriteString(e.Start.Format(time.RFC3339))
	buffer.WriteString(e.Stop.Format(time.RFC3339))
	buffer.WriteString(e.Notes)

	md5sum := md5.Sum(buffer.Bytes())

	return hex.EncodeToString(md5sum[:])
}
