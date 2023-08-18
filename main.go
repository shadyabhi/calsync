package main

import (
	"calsync/calendar/gcal"
	"calsync/maccalendar"
	"context"
	"fmt"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

func main() {
	ctx := context.Background()

	gCli, err := newGoogleClient(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize Google client: %s", err)
	}

	macCli, err := maccalendar.New(ctx, "Calendar")
	if err != nil {
		log.Fatalf("Failed to initialize Mac's client: %s", err)
	}

	events, err := macCli.GetEvents()
	if err != nil {
		log.Fatalf("Couldn't get list of events from Mac calendar: %s", err)
	}

	gCli.DeleteAll()
	gCli.PublishAll(events)
}

func newGoogleClient(ctx context.Context) (*gcal.Client, error) {
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	client, err := gcal.New(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to get initialize gcal: %s", err)
	}

	return client, nil
}
