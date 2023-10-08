package main

import (
	"calsync/calendar/gcal"
	"calsync/calendar/maccalendar"
	"calsync/config"
	"context"
	"fmt"
	"log"
	"os"

	"calsync/calendar"

	"golang.org/x/oauth2/google"
	gCalenader "google.golang.org/api/calendar/v3"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	log.Printf("Running calsync version: %s-%s-%s", version, commit, date)

	ctx := context.Background()

	cfg, err := config.GetConfig(os.Getenv("HOME") + "/.config/calsync/config.toml")
	if err != nil {
		log.Fatalf("Failed to get config: %s", err)
	}

	events, err := getMacEvents(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to get events from Mac calendar: %s", err)
	}

	gCli, err := newGoogleClient(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Google client: %s", err)
	}
	if err := gCli.SyncCalendar(events); err != nil {
		log.Fatalf("Failed to sync calendar: %s", err)
	}
}

func getMacEvents(ctx context.Context, cfg *config.Config) ([]calendar.Event, error) {
	macCli, err := maccalendar.New(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize Mac's client: %s", err)
	}
	events, err := macCli.GetEvents()
	if err != nil {
		return nil, fmt.Errorf("Couldn't get list of events from Mac calendar: %s", err)
	}

	return events, nil

}

func newGoogleClient(ctx context.Context, cfg *config.Config) (*gcal.Client, error) {
	b, err := os.ReadFile(cfg.CredentialsFile())
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file, location: %s: err: %w", cfg.CredentialsFile(), err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	oAuthCfg, err := google.ConfigFromJSON(b, gCalenader.CalendarScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to oAuthCfg: %v", err)
	}

	client, err := gcal.New(ctx, cfg, oAuthCfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to get initialize gcal: %s", err)
	}

	return client, nil
}
