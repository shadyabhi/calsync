package maccalendar

import (
	"calsync/calendar"
	"calsync/config"
	"context"
	"fmt"
)

type Calendar struct {
	ctx             context.Context
	iCalBuddyBinary string
	calName         string
	nDays           int
}

func New(ctx context.Context, cfg *config.Config) (*Calendar, error) {
	return &Calendar{
		ctx:             ctx,
		iCalBuddyBinary: cfg.Mac.ICalBuddyBinary,
		calName:         cfg.Mac.Name,
		nDays:           cfg.Mac.Days,
	}, nil
}

func (c *Calendar) GetEvents() ([]calendar.Event, error) {
	events, err := getEvents(c.iCalBuddyBinary, c.calName, c.nDays)
	if err != nil {
		return nil, fmt.Errorf("getting events from mac calendar: %s", err)
	}

	return events, nil
}
