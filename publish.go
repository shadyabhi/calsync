package main

import (
	"fmt"
	"log"
	"time"

	"google.golang.org/api/calendar/v3"
)

func publishEvent(srv *calendar.Service, event Event) error {
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

	calEntry, err := srv.Events.Insert(workCalID, calEntry).Do()
	if err != nil {
		return err
	}

	log.Printf("Event created: %s\n", calEntry.HtmlLink)

	return nil
}
