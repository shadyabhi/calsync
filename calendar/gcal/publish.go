package gcal

import (
	"calsync/calendar"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"golang.org/x/exp/slices"
	"golang.org/x/oauth2"
	googlecalendar "google.golang.org/api/calendar/v3"
)

const EventSourceTitle = "calsync"

// SyncToDest will sync all events to Google Calendar
// - If event is present in Google Calendar but not in local calendar, it will be deleted
func (c *Client) SyncToDest(calEvents []calendar.Event) error {
	start := time.Now()

	calendar.Events(calEvents).SortStartTime()

	eventsFromGoogle, err := c.GetAllGCalEvents(calEvents[0].Start, calEvents[len(calEvents)-1].Stop)
	if err != nil {
		return fmt.Errorf("Getting all events failed: %w", err)
	}

	slog.Info("Starting syncing all events", "total_local_events", len(calEvents), "total_gcal_events", len(eventsFromGoogle))

	dupFinder := newDuplicateEventsFinder()
	foundIndicesCalEvents := make([]int, 0)
	// Clean up stale events at Google Calendar
	for _, event := range eventsFromGoogle {
		// Events already created, skip them
		exists, position := dupFinder.isGCalinEvents(event, calEvents)
		if exists {
			slog.Info("Already synced", "summary", event.Summary, "start", event.Start.DateTime, "end", event.End.DateTime)
			foundIndicesCalEvents = append(foundIndicesCalEvents, position)
			continue
		}

		// Exists in Google, but not local calendar, time to delete
		if event.Source != nil && event.Source.Title == EventSourceTitle {
			slog.Info("Stale, deleting", "summary", event.Summary, "start", event.Start.DateTime, "end", event.End.DateTime)
			if err := c.Svc.Events.Delete(c.workCalID, event.Id).Do(); err != nil {
				return fmt.Errorf("Cleanup up existing event failed: %w", err)
			}
		} else {
			// Manually created event, not via calsync, leave it alone!
			slog.Info("Skipped deletion: this is not calsync managed", "summary", event.Summary, "start", event.Start.DateTime, "end", event.End.DateTime)
		}
	}

	// Create new events, as needed
	for i, event := range calEvents {
		if slices.Contains(foundIndicesCalEvents, i) {
			// Event already exists, skip it
			continue
		}
		if err := c.publishEvent(event); err != nil {
			return fmt.Errorf("publishing event: %s ,%w", event, err)
		}
	}

	slog.Info("Finished syncing all events to Google", "duration", time.Since(start))

	return nil
}

func (c *Client) GetAllGCalEvents(start time.Time, end time.Time) ([]*Event, error) {
	slog.Info("Start getting all events...")

	gEvents, err := c.Svc.Events.List(c.workCalID).
		ShowDeleted(false).
		SingleEvents(true).
		// 2500 is the max possible from API
		MaxResults(2500).
		TimeMin(start.Format(time.RFC3339)).
		TimeMax(end.Format(time.RFC3339)).
		OrderBy("startTime").
		Do()
	if err != nil {
		// TODO: Better error checking, why does errors.Is/As doesn't work here?
		unwrapped := errors.Unwrap(err)
		if _, ok := unwrapped.(*oauth2.RetrieveError); ok {
			if strings.Contains(unwrapped.Error(), "unauthorized_client") {
				return nil, fmt.Errorf("Invalid token.json file, got oauth unauthorized_client error: %w: %w", err, ErrInvalidToken)
			}
		}
		if strings.Contains(unwrapped.Error(), "Not Found") {
			// c.workCalID doesn't exist
			slog.Error("Configured Google Calendar doesn't exist on this account", "workCalID", c.workCalID)
			calList, err := c.Svc.CalendarList.List().Do()
			if err != nil {
				slog.Error("Couldn't get list of calendars")
				os.Exit(1)
			}
			allExistingIDs := make([]string, len(calList.Items))
			for _, cal := range calList.Items {
				allExistingIDs = append(allExistingIDs, cal.Id)
			}
			return nil, fmt.Errorf("configured workCalID doesn't exist, got: %s, all existing IDs: %v: %w", c.workCalID, allExistingIDs, ErrCalendarNotFound)
		}

		slog.Error("Unable to retrieve events from Google, unhandled error", "error", err)
		os.Exit(1)
	}

	events := make([]*Event, 0)
	for _, event := range gEvents.Items {
		events = append(events, &Event{event})
	}

	return events, nil
}

// PublishAllEvents unconditionally publishes new events to Google Calendar, without checking if they already exist
func (c *Client) PublishAllEvents(events []calendar.Event) error {
	start := time.Now()

	for _, event := range events {
		if err := c.publishEvent(event); err != nil {
			return fmt.Errorf("Publishing the event failed: event: %s , %w", event, err)
		}
	}

	slog.Info("Finished adding all events to Google", "duration", time.Since(start))

	return nil
}

// // DeleteAllEvents unconditionally deletes all events from Google Calendar
// func (c *Client) DeleteAllEvents(endTime time.Time) error {
// 	start := time.Now()
// 	log.Printf("Starting deletion of existing events...")

// 	eventsFromGoogle, err := c.GetAllGCalEvents(endTime)
// 	if err != nil {
// 		return fmt.Errorf("Getting all events failed: %w", err)
// 	}

// 	for _, event := range eventsFromGoogle {
// 		// Manually created event, not via calsync, leave it alone!
// 		if event.Source.Title != EventSourceTitle {
// 			log.Printf("Skipping deletion of event: %s", event.Summary)
// 		}

// 		if err := c.Svc.Events.Delete(c.workCalID, event.Id).Do(); err != nil {
// 			return fmt.Errorf("Cleanup up existing elements failed: %w", err)
// 		} else {
// 			log.Printf("Successfully deleted event: %s: %s %s", event.Summary, event.Start.DateTime, event.End.DateTime)
// 		}
// 	}

// 	log.Printf("Finished deleting existing events in %s", time.Since(start))

// 	return nil
// }

func (c *Client) publishEvent(event calendar.Event) error {
	calEntry := &googlecalendar.Event{
		Summary:     event.Title,
		Description: event.Notes,
		Start: &googlecalendar.EventDateTime{
			DateTime: event.Start.Format(time.RFC3339),
		},
		End: &googlecalendar.EventDateTime{
			DateTime: event.Stop.Format(time.RFC3339),
		},
		Source: &googlecalendar.EventSource{
			Title: "calsync",
			Url:   "https://github.com/shadyabhi/calsync",
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

	slog.Info("Event created", "summary", calEntry.Summary, "start", calEntry.Start.DateTime, "end", calEntry.End.DateTime)

	return nil
}
