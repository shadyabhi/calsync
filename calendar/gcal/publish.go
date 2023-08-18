package gcal

import (
	"calsync/calendar"
	"fmt"
	"log"
	"time"

	googlecalendar "google.golang.org/api/calendar/v3"
)

func (c *Client) PublishAll(events []calendar.Event) {
	start := time.Now()

	for _, event := range events {
		if err := c.PublishEvent(event); err != nil {
			log.Fatalf("Publishing all events failed: event: %s , %s", event, err)
		}
	}

	log.Printf("Finished adding all events to Google in %s", time.Since(start))
}

func (c *Client) PublishEvent(event calendar.Event) error {
	calEntry := &googlecalendar.Event{
		Summary: event.Title,
		Start: &googlecalendar.EventDateTime{
			DateTime: event.Start.Format(time.RFC3339),
		},
		End: &googlecalendar.EventDateTime{
			DateTime: event.Stop.Format(time.RFC3339),
		},
		Description: fmt.Sprintf("uid=%s", event.UID),
	}

	calEntry, err := c.Svc.Events.Insert(WorkCalID, calEntry).Do()
	if err != nil {
		return err
	}

	log.Printf("Event created: %s, %s, %s\n", calEntry.Summary, calEntry.Start.DateTime, calEntry.End.DateTime)

	return nil
}
