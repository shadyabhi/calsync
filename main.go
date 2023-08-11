package main

import (
	"calsync/gcal"
	"context"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := gcal.GetClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	deleteExisting(srv)
	publishAll(srv)
}

func publishAll(srv *calendar.Service) {
	start := time.Now()
	events, err := getEvents()
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		if err := PublishEvent(srv, event); err != nil {
			log.Fatalf("Publishing all events failed: event: %s , %s", event, err)
		}
	}
	log.Printf("Finished adding all events to Google in %s", time.Since(start))

}

func deleteExisting(srv *calendar.Service) {
	log.Printf("Starting deleting of existing events...")
	start := time.Now()
	// Hack: Truncate works with UTC, so we need to include the whole day
	t := time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)
	eventsFromGoogle, err := srv.Events.List(gcal.WorkCalID).ShowDeleted(false).
		SingleEvents(true).TimeMin(t.Format(time.RFC3339)).TimeMax(t.Add(14 * 24 * time.Hour).Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve events from Google: %v", err)
	}
	for _, event := range eventsFromGoogle.Items {
		if err := srv.Events.Delete(gcal.WorkCalID, event.Id).Do(); err != nil {
			log.Fatalf("Cleanup up existing elements failed: %s", err)
		}
	}
	log.Printf("Finished deleting existing events in %s", time.Since(start))
}
