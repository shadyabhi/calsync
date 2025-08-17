package maccalendar

import (
	"calsync/calendar"
	"calsync/config"
	"context"
	"fmt"
	"time"
)

type Calendar struct {
	ctx             context.Context
	iCalBuddyBinary string
	calName         string
}

func New(ctx context.Context, cfg *config.Mac) (*Calendar, error) {
	return &Calendar{
		ctx:             ctx,
		iCalBuddyBinary: cfg.ICalBuddyBinary,
		calName:         cfg.Name,
	}, nil
}

func (c *Calendar) String() string {
	return fmt.Sprintf("Mac Calendar: %s", c.calName)
}

func (c *Calendar) GetEvents(start time.Time, end time.Time) ([]calendar.Event, error) {
	events, err := getEvents(c.iCalBuddyBinary, c.calName, start, end)
	if err != nil {
		return nil, fmt.Errorf("getting events from mac calendar: %s", err)
	}

	return events, nil
}

func (c *Calendar) DeleteAll(_ int) error { return nil }
func (c *Calendar) PutEvents() error      { return nil }
func (c *Calendar) SyncToDest([]calendar.Event) error {
	return fmt.Errorf("SyncToDest not implemented for Mac calendar")
}
