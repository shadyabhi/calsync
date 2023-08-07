package main

import (
	"os/exec"
	"strings"
	"time"
)

const (
	timeLayout = "Jan 2, 2006 15:04 -0700"
)

type Event struct {
	Title       string
	Start, Stop time.Time
	UID         string
}

func getSourceRaw() (string, error) {
	cmd := exec.Command("icalBuddy", []string{
		"-b",
		"---",
		"-eep", "notes,attendees,location",
		"-uid",
		"-ic", "Calendar",
		"-nc",
		"-tf", "%H:%M %z",
		"eventsToday+3",
	}...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func getEvents() ([]Event, error) {
	output, err := getSourceRaw()
	if err != nil {
		return nil, err
	}

	events := make([]Event, 0)

	for _, multilineEvent := range strings.Split(output, "---") {
		if len(multilineEvent) == 0 {
			continue
		}

		event, err := getEvent(multilineEvent)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func getEvent(raw string) (Event, error) {
	lines := strings.Split(raw, "\n")

	event := Event{
		Title: lines[0],
	}

	timeLine := lines[1]

	// 4 spaces to be deleted, all fields start with 4 spaces
	timeLine = timeLine[4:]

	atLocation := strings.Index(timeLine, " at ")
	datePart := normalizeDay(timeLine[:atLocation])

	timeParts := strings.Split(timeLine[atLocation+4:], " - ")

	parsedTime, err := time.Parse(timeLayout, datePart+" "+timeParts[0])
	if err != nil {
		return event, err
	}
	event.Start = parsedTime.In(time.Local)

	parsedTime, err = time.Parse(timeLayout, datePart+" "+timeParts[1])
	if err != nil {
		return event, err
	}
	event.Stop = parsedTime.In(time.Local)

	uidLine := lines[2]
	event.UID = uidLine[9:]

	return event, nil
}

// normalizeDay normalizes the date when it's not a number
// Example:-
// today
// tomorrow
// day after tomorrow
func normalizeDay(date string) string {
	now := time.Now().Local()

	switch date {
	case "today":
		return now.Format("Jan 2, 2006")
	case "tomorrow":
		return now.Add(24 * time.Hour).Format("Jan 2, 2006")
	case "day after tomorrow":
		return now.Add(48 * time.Hour).Format("Jan 2, 2006")
	default:
		return date
	}
}
