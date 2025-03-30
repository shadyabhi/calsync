package ics

import (
	"calsync/calendar"
	"calsync/config"
	"context"
	"fmt"
	"net/http"
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

	var tzMapping = map[string]string{
		"Pacific Standard Time": "America/Los_Angeles",
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
		event := calendar.Event{}
		event.Title = sourceEvent.Summary

		event.Start = *sourceEvent.Start
		event.Stop = *sourceEvent.End

		event.UID = sourceEvent.Uid

		events = append(events, event)
	}

	return events, nil
}
