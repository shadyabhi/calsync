package cmd

import (
	"calsync/calendar"
	"calsync/calendar/gcal"
	"calsync/calendar/ics"
	"calsync/calendar/maccalendar"
	"calsync/config"
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"time"

	"golang.org/x/oauth2/google"
	gCalenader "google.golang.org/api/calendar/v3"
)

func syncCalendars(ctx context.Context, cfg *config.Config) {
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

func getCalendarsFor(ctx context.Context, cfg *config.Config, typ config.Calendars) ([]calendar.Calendar, error) {
	var calendars []calendar.Calendar

	v := reflect.ValueOf(typ)

	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i).Interface()

		cal, err := initializeCalendar(ctx, cfg, fieldValue)
		if err != nil {
			return nil, fmt.Errorf("Failed to initialize calendar client: %w", err)
		}
		if cal != nil {
			calendars = append(calendars, cal)
		}
	}

	return calendars, nil
}

func initializeCalendar(ctx context.Context, cfg *config.Config, fieldValue interface{}) (calendar.Calendar, error) {
	switch concrete := fieldValue.(type) {
	case *config.Mac:
		if concrete != nil && concrete.Enabled {
			return maccalendar.New(ctx, concrete)
		}
	case *config.ICal:
		if concrete != nil && concrete.Enabled {
			return ics.New(ctx, *concrete)
		}
	case *config.Google:
		if concrete != nil && concrete.Enabled {
			return newGoogleClient(ctx, cfg)
		}
	}
	return nil, nil
}

func getCalendarByName(ctx context.Context, cfg *config.Config, calendarName string) (calendar.Calendar, error) {
	// Check both source and target calendars
	for _, calendars := range []config.Calendars{cfg.Source, cfg.Target} {
		v := reflect.ValueOf(calendars)
		t := reflect.TypeOf(calendars)

		for i := 0; i < v.NumField(); i++ {
			fieldName := strings.ToLower(t.Field(i).Name)
			fieldValue := v.Field(i).Interface()

			// Match the calendar name with the field name
			if fieldName == calendarName {
				cal, err := initializeCalendar(ctx, cfg, fieldValue)
				if err != nil {
					return nil, fmt.Errorf("failed to initialize %s calendar: %w", calendarName, err)
				}
				if cal != nil {
					return cal, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("calendar %s not found or not enabled", calendarName)
}
