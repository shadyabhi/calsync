package main

import (
	"calsync/calendar/gcal"
	"calsync/calendar/ics"
	"calsync/calendar/maccalendar"
	"calsync/config"
	versioncheck "calsync/version"
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"time"

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
	// Setup slog with a format that resembles the current log output
	// Set to Debug level to enable all logs, can be controlled via environment variable
	level := slog.LevelInfo
	if os.Getenv("DEBUG") != "" {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove the time prefix and level prefix to match current format more closely
			if a.Key == slog.TimeKey || a.Key == slog.LevelKey {
				return slog.Attr{}
			}
			return a
		},
	}))
	slog.SetDefault(logger)

	slog.Info("Running calsync", "version", fmt.Sprintf("%s-%s-%s", version, commit, date))

	versionChan := make(chan *versioncheck.UpdateInfo, 1)
	versionChecker := versioncheck.New(version)
	// Start version check in background (non-blocking)
	go versionChecker.CheckForUpdate(versionChan)

	ctx := context.Background()

	cfg, err := config.GetConfig(os.Getenv("HOME") + "/.config/calsync/config.toml")
	if err != nil {
		slog.Error("Failed to get config", "error", err)
		os.Exit(1)
	}

	sources, targets, err := getSourceTargetCalendars(ctx, cfg)
	if err != nil {
		slog.Error("Failed to get source and target calendars", "error", err)
		os.Exit(1)
	}

	start := time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)
	end := start.Add(24 * time.Hour * time.Duration(cfg.Sync.Days))

	slog.Info("Searching for events", "start", start.Format(time.RFC3339), "end", end.Format(time.RFC3339))

	events, err := getSourceEventsSorted(ctx, sources, start, end)
	if err != nil {
		slog.Error("Failed to get events from Mac calendar", "error", err)
		os.Exit(1)
	}

	for _, target := range targets {
		if err := target.SyncToDest(events); err != nil {
			slog.Error("Failed to sync events to target calendar", "error", err)
			os.Exit(1)
		}
	}

	// Check for update info at the very end
	select {
	case update := <-versionChan:
		if update != nil && update.Available {
			slog.Warn("A newer version is available",
				"current", version,
				"latest", update.LatestVersion,
				"action", "Run 'brew upgrade calsync' to update")
		}
	default:
		slog.Debug("No update info received, this version is up-to-date")
	}
}

func getSourceEventsSorted(_ context.Context, sources []calendar.Calendar, start time.Time, end time.Time) ([]calendar.Event, error) {
	allEvents := make([]calendar.Event, 0)
	for _, src := range sources {
		events, err := src.GetEvents(start, end)
		if err != nil {
			return nil, fmt.Errorf("Couldn't get list of events from source calendar: %s", err)
		}
		allEvents = append(allEvents, events...)
	}

	calendar.Events(allEvents).SortStartTime()

	return allEvents, nil
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

	client, err := gcal.New(ctx, *cfg.Target.Google, oAuthCfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to get initialize gcal: %s", err)
	}

	return client, nil
}

func getSourceTargetCalendars(ctx context.Context, cfg *config.Config) ([]calendar.Calendar, []calendar.Calendar, error) {

	sources, err := getCalendarsFor(ctx, cfg, cfg.Source)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get source calendars: %s", err)
	}
	if len(sources) == 0 {
		return nil, nil, fmt.Errorf("no enabled source calendars found")
	}

	targets, err := getCalendarsFor(ctx, cfg, cfg.Target)
	if err != nil {
		return nil, nil, fmt.Errorf("Failed to get target calendars: %s", err)
	}
	if len(targets) == 0 {
		return nil, nil, fmt.Errorf("no enabled target calendars found")
	}

	slog.Info("Configured calendars", "sources", sources, "targets", targets)

	return sources, targets, nil

}

// getCalendarsFor returns a list of calendars based on the provided config.
// It uses reflection to iterate over the fields of the config struct and
// initializes the corresponding calendar clients if they are enabled.
// It returns a slice of calendar.Calendar interfaces.
//
// Configs are flexible to generate a runtime list of calendars, so reflection
// is opted here.
func getCalendarsFor(ctx context.Context, cfg *config.Config, typ config.Calendars) ([]calendar.Calendar, error) {
	var calendars []calendar.Calendar

	v := reflect.ValueOf(typ)
	values := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()

		switch concrete := values[i].(type) {
		case *config.Mac:
			if concrete != nil && concrete.Enabled {
				macCal, err := maccalendar.New(ctx, typ.Mac)
				if err != nil {
					return nil, fmt.Errorf("Failed to initialize Mac calendar client: %s", err)
				}
				calendars = append(calendars, macCal)
			}
		case *config.ICal:
			if concrete != nil && concrete.Enabled {
				icsClient, err := ics.New(ctx, *typ.ICal)
				if err != nil {
					return nil, fmt.Errorf("Failed to initialize ICS client: %s", err)
				}

				calendars = append(calendars, icsClient)
			}
		case *config.Google:
			if concrete != nil && concrete.Enabled {
				gCli, err := newGoogleClient(ctx, cfg)
				if err != nil {
					return nil, fmt.Errorf("Failed to initialize Google client: %s", err)
				}
				calendars = append(calendars, gCli)
			}
		}
	}

	return calendars, nil
}
