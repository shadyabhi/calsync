package maccalendar

import (
	"bufio"
	"calsync/calendar"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

var ErrFullDayEvent = fmt.Errorf("full day event")

const (
	timeLayout      = "Jan 2, 2006 15:04 -0700"
	iCalBulletPoint = "→"
)

func getEvents(iCalBuddyBinary string, calName string, nDays int) ([]calendar.Event, error) {
	output, err := getSourceRaw(iCalBuddyBinary, calName, nDays)
	if err != nil {
		return nil, fmt.Errorf("getting source raw: %s", err)
	}

	events := make([]calendar.Event, 0)

	for _, multilineEvent := range strings.Split(output, iCalBulletPoint) {
		log.Printf("Parsing event: \"%s\"\n", multilineEvent)

		if len(multilineEvent) == 0 {
			log.Printf("Skipped due to blank event")
			continue
		}

		event, err := getEvent(multilineEvent)
		if errors.Is(err, ErrFullDayEvent) {
			log.Printf("Skipped due to full day event")
			continue
		}

		if err != nil {
			return nil, fmt.Errorf("parsing event %s, got error %w", multilineEvent, err)
		}
		events = append(events, event)
	}

	return events, nil
}

func getSourceRaw(icalBuddyBinary string, calName string, nDays int) (string, error) {
	cmd := exec.Command(icalBuddyBinary, []string{
		"-b",
		iCalBulletPoint,
		"-eep", "attendees,location",
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

// TODO: Refactor to multiple methods
func getEvent(raw string) (calendar.Event, error) {
	event := calendar.Event{}
	reader := bufio.NewReader(strings.NewReader(raw))

	// Line 1: Title
	line, err := reader.ReadString('\n')
	if err != nil {
		return event, fmt.Errorf("reading title: %s", err)
	}
	event.Title = line[:len(line)-1]

	// Line 2: notes or time
	line, err = reader.ReadString('\n')
	if err != nil {
		return event, fmt.Errorf("reading notes: %s", err)
	}

	if strings.HasPrefix(line, "    notes: ") {
		notesBody := strings.TrimPrefix(line, "    notes: ")
		for {
			line, err = reader.ReadString('\n')
			if err != nil {
				return event, fmt.Errorf("reading notes body: %s", err)
			}
			// We're still reading notes body
			if strings.HasPrefix(line, "        ") {
				notesBody += line[4:]
			} else {
				break
			}
		}
		event.Notes = strings.TrimSuffix(notesBody, "\n")
	}

	// Time
	timeLine := line[4 : len(line)-1]

	atLocation := strings.Index(timeLine, " at ")
	if atLocation == -1 {
		log.Printf("Event: '%s' doesn't have time associated to it, skipping for now", raw)
		return event, ErrFullDayEvent
	}

	startDate := timeLine[:atLocation] + " "

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
		timeParts[1] = strings.Replace(timeParts[1], " at ", " ", 1)
	}

	parsedTime, err = time.Parse(timeLayout, endDate+timeParts[1])
	if err != nil {
		return event, fmt.Errorf("parsing stop time: %s", err)
	}
	event.Stop = parsedTime.In(time.Local)

	// Line: uid
	line, err = reader.ReadString('\n')
	if err != nil {
		return event, fmt.Errorf("reading uid: %s", err)
	}
	uidLine := line[:len(line)-1]
	event.UID = uidLine[9:]

	return event, nil

}
