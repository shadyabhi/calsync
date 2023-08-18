package gcal

import (
	"log"
	"time"
)

func (c *Client) DeleteAll() {
	start := time.Now()
	log.Printf("Starting deletion of existing events...")

	// Hack: Truncate works with UTC, so we need to include the whole day
	t := time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)
	eventsFromGoogle, err := c.Svc.Events.List(WorkCalID).ShowDeleted(false).
		SingleEvents(true).TimeMin(t.Format(time.RFC3339)).TimeMax(t.Add(14 * 24 * time.Hour).Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve events from Google: %v", err)
	}

	for _, event := range eventsFromGoogle.Items {
		if err := c.Svc.Events.Delete(WorkCalID, event.Id).Do(); err != nil {
			log.Fatalf("Cleanup up existing elements failed: %s", err)
		}
	}

	log.Printf("Finished deleting existing events in %s", time.Since(start))

}
