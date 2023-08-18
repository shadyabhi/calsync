package maccalendar

import (
	"calsync/calendar"
	"context"
	"fmt"
)

type Calendar struct {
	ctx     context.Context
	calName string
}

func New(ctx context.Context, calName string) (*Calendar, error) {
	return &Calendar{
		ctx:     ctx,
		calName: calName,
	}, nil
}

func (c *Calendar) GetEvents() ([]calendar.Event, error) {

	events, err := getEvents()
	if err != nil {
		return nil, fmt.Errorf("getting events from mac calendar: %s", err)
	}

	return events, nil
}
