package maccalendar

import (
	"calsync/calendar"
	"log"
	"os/exec"
	"strings"
	"time"
)

func getEvents() ([]calendar.Event, error) {
	output, err := getSourceRaw()
	if err != nil {
		return nil, err
	}

	events := make([]calendar.Event, 0)

	for _, multilineEvent := range strings.Split(output, "---") {
		log.Printf("Parsing event: \"%s\"\n", multilineEvent)

		if len(multilineEvent) == 0 {
			log.Printf("Skipped due to blank event")
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

const (
	timeLayout = "Jan 2, 2006 15:04 -0700"
)

func getSourceRaw() (string, error) {
	cmd := exec.Command("icalBuddy", []string{
		"-b",
		"---",
		"-eep", "notes,attendees,location",
		"-uid",
		"-ic", "Calendar",
		"-nc",
		"-tf", "%H:%M %z",
		"eventsToday+7",
	}...)
	log.Printf("Running icalBuddy with args: %s", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}

func getEvent(raw string) (calendar.Event, error) {
	lines := strings.Split(raw, "\n")

	event := calendar.Event{
		Title: lines[0],
	}

	timeLine := lines[1]

	// 4 spaces to be deleted, all fields start with 4 spaces
	timeLine = timeLine[4:]

	atLocation := strings.Index(timeLine, " at ")
	if atLocation == -1 {
		log.Printf("Event: '%s' doesn't have time associated to it, skipping for now", raw)
		return event, nil
	}

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
