package gcal

import (
	"calsync/calendar"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2"
	googlecalendar "google.golang.org/api/calendar/v3"
)

func (c *Client) PublishAllEvents(events []calendar.Event) {
	start := time.Now()

	for _, event := range events {
		if err := c.PublishEvent(event); err != nil {
			log.Fatalf("Publishing all events failed: event: %s , %s", event, err)
		}
	}

	log.Printf("Finished adding all events to Google in %s", time.Since(start))
}

// SyncCalendar will sync all events to Google Calendar, also remove any extra events.
func (c *Client) SyncCalendar(events []calendar.Event) {
	start := time.Now()

	/*
		- Get all events from Google
		- Use UID in extendedProperties to match events with local variable
		- If UID is not found locally, delete from Google Calendar
		- If UID is found locally, update Google Calendar
	*/

	for _, event := range events {
		if err := c.SyncEvent(event); err != nil {
			log.Fatalf("Syncing all events failed: event: %s , %s", event, err)
		}
	}

	log.Printf("Finished syncing all events to Google in %s", time.Since(start))
}

func (c *Client) SyncEvent(event calendar.Event) error {
	return nil
}

func (c *Client) GetAllGCalEvents() ([]*googlecalendar.Event, error) {
	start := time.Now()
	log.Printf("Start getting all events...")

	// Hack: Truncate works with UTC, so we need to include the whole day
	// If we run script at 01:00, we shouldn't miss events from 00:00-01:00
	startTimeMidnight := time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)
	eventsFromGoogle, err := c.Svc.Events.List(c.workCalID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(startTimeMidnight.Format(time.RFC3339)).
		TimeMax(startTimeMidnight.Add(14 * 24 * time.Hour).Format(time.RFC3339)).
		OrderBy("startTime").
		Do()
	if err != nil {
		// TODO: Better error checking
		unwrapped := errors.Unwrap(err)
		if _, ok := unwrapped.(*oauth2.RetrieveError); ok {
			if strings.Contains(unwrapped.Error(), "unauthorized_client") {
				return nil, fmt.Errorf("Invalid token.json file, got oauth unauthorized_client error: %w: %w", err, ErrInvalidToken)
			}
		}
		if strings.Contains(unwrapped.Error(), "Not Found") {
			// c.workCalID doesn't exist
			log.Printf("Configured Google Calendar doesn't exist on this account: %s", c.workCalID)
			calList, err := c.Svc.CalendarList.List().Do()
			if err != nil {
				log.Fatalf("Couldn't get list of calendars")
			}
			allExistingIDs := make([]string, len(calList.Items))
			for _, cal := range calList.Items {
				allExistingIDs = append(allExistingIDs, cal.Id)
			}
			return nil, fmt.Errorf("configured workCalID doesn't exist, got: %s, all existing IDs: %v: %w", c.workCalID, allExistingIDs, ErrCalendarNotFound)
		}

		log.Fatalf("Unable to retrieve events from Google, unhandled error: %v", err)
	}

	log.Printf("Finished getting all events in %s", time.Since(start))

	return eventsFromGoogle.Items, nil
}

func (c *Client) DeleteAllEvents() error {
	start := time.Now()
	log.Printf("Starting deletion of existing events...")

	eventsFromGoogle, err := c.GetAllGCalEvents()
	if err != nil {
		return fmt.Errorf("Getting all events failed: %w", err)
	}

	for _, event := range eventsFromGoogle {
		// Manually created event, not via calsync, leave it alone!
		if event.Source.Title != "calsync" {
			log.Printf("Skipping deletion of event: %s", event.Summary)
		}

		if err := c.Svc.Events.Delete(c.workCalID, event.Id).Do(); err != nil {
			return fmt.Errorf("Cleanup up existing elements failed: %w", err)
		} else {
			log.Printf("Successfully deleted event: %s: %s %s", event.Summary, event.Start.DateTime, event.End.DateTime)
		}
	}

	log.Printf("Finished deleting existing events in %s", time.Since(start))

	return nil
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
		Source: &googlecalendar.EventSource{
			Title: "calsync",
			Url:   "https://calsync.local",
		},
		ExtendedProperties: &googlecalendar.EventExtendedProperties{
			Private: map[string]string{
				"uid": event.UID,
			},
		},
	}

	calEntry, err := c.Svc.Events.Insert(c.workCalID, calEntry).Do()
	if err != nil {
		return err
	}

	log.Printf("Event created: %s, %s, %s\n", calEntry.Summary, calEntry.Start.DateTime, calEntry.End.DateTime)

	return nil
}
