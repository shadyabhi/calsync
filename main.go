package main

import (
	"calsync/calendars/gcal"
	"context"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
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

	calendar, err := gcal.New(ctx, config)
	if err != nil {
		log.Fatalf("Failed to get initialize gcal: %s", err)
	}

	calendar.DeleteAll()
	calendar.PublishAll()
}
