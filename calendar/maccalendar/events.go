package maccalendar

import (
	"calsync/calendar"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

func getEvents(iCalBuddyBinary string, calName string, nDays int) ([]calendar.Event, error) {
	output, err := getSourceRaw(iCalBuddyBinary, calName, nDays)
	if err != nil {
		return nil, fmt.Errorf("getting source raw: %s", err)
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
			return nil, fmt.Errorf("parsing event %s, got error %w", multilineEvent, err)
		}
		events = append(events, event)
	}

	return events, nil
}

const (
	timeLayout = "Jan 2, 2006 15:04 -0700"
)

func getSourceRaw(icalBuddyBinary string, calName string, nDays int) (string, error) {
	cmd := exec.Command(icalBuddyBinary, []string{
		"-b",
		"---",
		"-eep", "notes,attendees,location",
		"-uid",
		"-ic", calName,
		"-nc",
		"-tf", "%H:%M %z",
		"-nrd",
		fmt.Sprintf("eventsToday+%d", nDays),
	}...)
	log.Printf("Running icalBuddy with args: %s", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running icalBuddy, output, %s: error: %s", output, err)
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

	startDate := normalizeDay(timeLine[:atLocation]) + " "

	// Split start and stop time
	timeParts := strings.Split(timeLine[atLocation+4:], " - ")

	parsedTime, err := time.Parse(timeLayout, startDate+timeParts[0])
	if err != nil {
		return event, fmt.Errorf("parsing start time: %s", err)
	}
	event.Start = parsedTime.In(time.Local)

	var endDate string
	if len(timeParts[1]) <= 11 {
		// Time without date
		endDate = startDate
	} else {
		// time has "at"
		// First part is date: <date> at <time>
		// or
		// First part is: "tomorrow at <time>"

		timeParts[1] = normalizeDay(timeParts[1])
		timeParts[1] = strings.Replace(timeParts[1], " at ", " ", 1)
	}

	parsedTime, err = time.Parse(timeLayout, endDate+timeParts[1])
	if err != nil {
		return event, fmt.Errorf("parsing stop time: %s", err)
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

	commonDatePattern := "day after tomorrow"
	if strings.Contains(date, commonDatePattern) {
		afterDate := date[strings.Index(date, commonDatePattern)+len(commonDatePattern):]
		return now.Add(48*time.Hour).Format("Jan 2, 2006") + afterDate
	}

	commonDatePattern = "tomorrow"
	if strings.Contains(date, commonDatePattern) {
		afterDate := date[strings.Index(date, commonDatePattern)+len(commonDatePattern):]
		return now.Add(24*time.Hour).Format("Jan 2, 2006") + afterDate
	}

	commonDatePattern = "today"
	if strings.Contains(date, commonDatePattern) {
		afterDate := date[strings.Index(date, commonDatePattern)+len(commonDatePattern):]
		return now.Format("Jan 2, 2006") + afterDate
	}

	// If none of those phrases are found, we return date, as is.
	return date
}
