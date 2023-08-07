package main

import (
	"fmt"
	"log"
	"time"

	"github.com/davecgh/go-spew/spew"
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
	spew.Dump(calEntry)

	log.Printf("Will add the following event: %#v", event)

	calEntry, err := srv.Events.Insert(workCalID, calEntry).Do()
	if err != nil {
		return err
	}

	fmt.Printf("Event created: %s\n", calEntry.HtmlLink)
	spew.Dump(calEntry)

	return nil
}
