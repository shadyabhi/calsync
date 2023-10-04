package gcal

import (
	"errors"
	"log"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

func (c *Client) DeleteAll() {
	start := time.Now()
	log.Printf("Starting deletion of existing events...")

	// Hack: Truncate works with UTC, so we need to include the whole day
	t := time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)
	eventsFromGoogle, err := c.Svc.Events.List(c.workCalID).ShowDeleted(false).
		SingleEvents(true).TimeMin(t.Format(time.RFC3339)).TimeMax(t.Add(14 * 24 * time.Hour).Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		// TODO: Better error checking
		unwrapped := errors.Unwrap(err)
		if _, ok := unwrapped.(*oauth2.RetrieveError); ok {
			if strings.Contains(unwrapped.Error(), "unauthorized_client") {
				log.Printf("Invalid token.json file, got oauth unauthorized_client error: %v", err)
				log.Fatalf("Please delete token.json file and try again, location: %s", c.cfg.TokenFile())
			}
		}
		if strings.Contains(unwrapped.Error(), "Not Found") {
			log.Printf("Configured Google Calendar doesn't exist on this account: %s", c.workCalID)
			calList, err := c.Svc.CalendarList.List().Do()
			if err != nil {
				log.Fatalf("Couldn't get list of calendars")
			}
			for i, cal := range calList.Items {
				log.Printf("Found following calendars on account %d: name: %s, id: %s", i, cal.Summary, cal.Id)
			}
			log.Fatalf("Create correct calendar at Google.Id at: %s, if you've not created one, visit: %s to create one",
				c.cfg.ConfigFile(),
				"https://calendar.google.com/calendar/r/settings/createcalendar")
		}

		log.Fatalf("Unable to retrieve events from Google: %v", err)
	}

	for _, event := range eventsFromGoogle.Items {
		// Manually created event, not via calsync, leave it alone!
		if event.Source.Title != "calsync" {
			log.Printf("Skipping deletion of event: %s", event.Summary)
		}

		if err := c.Svc.Events.Delete(c.workCalID, event.Id).Do(); err != nil {
			log.Fatalf("Cleanup up existing elements failed: %s", err)
		} else {
			log.Printf("Successfully deleted event: %s: %s %s", event.Summary, event.Start.DateTime, event.End.DateTime)
		}
	}

	log.Printf("Finished deleting existing events in %s", time.Since(start))

}
