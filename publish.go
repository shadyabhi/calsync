package main

import (
	"calsync/gcal"
	"fmt"
	"log"
	"time"

	"google.golang.org/api/calendar/v3"
)

func PublishEvent(srv *calendar.Service, event Event) error {
	calEntry := &calendar.Event{
		Summary: event.Title,
		Start: &calendar.EventDateTime{
			DateTime: event.Start.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: event.Stop.Format(time.RFC3339),
		},
		Description: fmt.Sprintf("uid=%s", event.UID),
	}

	calEntry, err := srv.Events.Insert(gcal.WorkCalID, calEntry).Do()
	if err != nil {
		return err
	}

	log.Printf("Event created: %s, %s, %s\n", calEntry.Summary, calEntry.Start.DateTime, calEntry.End.DateTime)

	return nil
}
