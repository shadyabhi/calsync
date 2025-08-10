package ics

import (
	"calsync/calendar"
	"calsync/config"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	gocal "github.com/apognu/gocal"
)

type Calendar struct {
	ctx context.Context
	url string
}

func New(ctx context.Context, cfg config.ICal) (*Calendar, error) {
	return &Calendar{
		ctx: ctx,
		url: cfg.URL,
	}, nil
}

func (c *Calendar) GetEvents(start time.Time, end time.Time) ([]calendar.Event, error) {
	events, err := getEvents(c.url, start, end)
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (c *Calendar) DeleteAll() error {
	return fmt.Errorf("DeleteAll not implemented for ICS calendar")
}

func (c *Calendar) PutEvents() error {
	// Not implemented
	return fmt.Errorf("PutEvents not implemented for ICS calendar")
}

func (c *Calendar) SyncToDest([]calendar.Event) error {
	return fmt.Errorf("SyncToDest not implemented for ICS calendar")
}

func getEvents(url string, start time.Time, end time.Time) ([]calendar.Event, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch calendar: %w", err)
	}
	defer resp.Body.Close()

	// https://github.com/unicode-org/cldr/blob/main/common/supplemental/windowsZones.xml
	var tzMapping = map[string]string{
		// US Timezones
		"Pacific Standard Time":  "America/Los_Angeles",
		"Mountain Standard Time": "America/Denver",
		"Central Standard Time":  "America/Chicago",
		"Eastern Standard Time":  "America/New_York",
		"Alaskan Standard Time":  "America/Anchorage",
		"Hawaiian Standard Time": "Pacific/Honolulu",

		// European Timezones
		"GMT Standard Time":              "Europe/London",
		"Greenwich Standard Time":        "Atlantic/Reykjavik",
		"W. Europe Standard Time":        "Europe/Berlin",
		"Central Europe Standard Time":   "Europe/Budapest",
		"Romance Standard Time":          "Europe/Paris",
		"Central European Standard Time": "Europe/Warsaw",
		"E. Europe Standard Time":        "Europe/Chisinau",
		"FLE Standard Time":              "Europe/Kiev",
		"Russian Standard Time":          "Europe/Moscow",

		// Asia-Pacific Timezones
		"China Standard Time":     "Asia/Shanghai",
		"Tokyo Standard Time":     "Asia/Tokyo",
		"Korea Standard Time":     "Asia/Seoul",
		"Singapore Standard Time": "Asia/Singapore",
		"India Standard Time":     "Asia/Calcutta",
		"Arabian Standard Time":   "Asia/Riyadh",
		"Iran Standard Time":      "Asia/Tehran",
		"Israel Standard Time":    "Asia/Jerusalem",

		// Australia/New Zealand
		"AUS Eastern Standard Time":  "Australia/Sydney",
		"AUS Central Standard Time":  "Australia/Darwin",
		"W. Australia Standard Time": "Australia/Perth",
		"New Zealand Standard Time":  "Pacific/Auckland",

		// Americas
		"Canada Central Standard Time": "America/Regina",
		"Mexico Standard Time":         "America/Mexico_City",
		"US Mountain Standard Time":    "America/Phoenix",
		"Atlantic Standard Time":       "America/Halifax",
		"Argentina Standard Time":      "America/Buenos_Aires",
		"Brazil Standard Time":         "America/Sao_Paulo",
		"Chile Standard Time":          "America/Santiago",

		// Africa/Middle East
		"South Africa Standard Time": "Africa/Johannesburg",
		"Egypt Standard Time":        "Africa/Cairo",
		"West Africa Standard Time":  "Africa/Lagos",
		"Middle East Standard Time":  "Asia/Beirut",

		// Common UTC variations
		"UTC":                        "UTC",
		"Coordinated Universal Time": "UTC",
		"GMT":                        "UTC",
	}

	gocal.SetTZMapper(func(s string) (*time.Location, error) {
		if tzid, ok := tzMapping[s]; ok {
			return time.LoadLocation(tzid)
		}
		return nil, fmt.Errorf("")
	})

	c := gocal.NewParser(resp.Body)
	c.Start, c.End = &start, &end
	if err := c.Parse(); err != nil {
		return nil, fmt.Errorf("failed to parse calendar: %w", err)
	}

	events := make([]calendar.Event, 0)

	for _, sourceEvent := range c.Events {
		slog.Debug("Processing ICS event",
			"uid", sourceEvent.Uid,
			"summary", sourceEvent.Summary,
			"start", sourceEvent.Start,
			"end", sourceEvent.End,
			"timezone", sourceEvent.RawStart.Params["TZID"])

		// Outlook's timezones don't follow the standard
		gotTZ := sourceEvent.RawStart.Params["TZID"]
		if isUnknownTZ(tzMapping, gotTZ) {
			// Timezone not found in mapping is a hard error, we abort!
			slog.Error("Timezone not found in mapping, using UTC", "timezone", gotTZ)
			os.Exit(1)
		}

		event := calendar.Event{}
		event.Title = sourceEvent.Summary

		event.Start = *sourceEvent.Start
		event.Stop = *sourceEvent.End

		event.UID = sourceEvent.Uid

		events = append(events, event)
	}

	return events, nil
}

// isUnknownTZ checks if the timezone is not found in the mapping
func isUnknownTZ(tzMapping map[string]string, gotTZ string) bool {
	var found bool
	for k := range tzMapping {
		if gotTZ == k {
			found = true
			break
		}
	}
	return !found
}
