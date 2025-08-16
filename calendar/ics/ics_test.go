package ics

import (
	"calsync/calendar"
	"calsync/config"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func serveICSFile(filename string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		file, err := os.Open(filename)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		defer file.Close()
		w.Header().Set("Content-Type", "text/calendar")
		_, err = io.Copy(w, file)
		if err != nil {
			http.Error(w, "Failed to read ICS file", http.StatusInternalServerError)
		}
	}))
}

func TestGetEvents(t *testing.T) {
	server := serveICSFile("testdata/test.ics")
	defer server.Close()

	ctx := context.Background()
	cfg := config.ICal{URL: server.URL}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	start := time.Date(2005, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2005, 8, 31, 23, 59, 59, 0, time.UTC)

	events, err := cal.GetEvents(start, end)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	expectedTitle := "Data::ICal release party,other things with slash\\es"
	if event.Title != expectedTitle {
		t.Errorf("Expected title '%s', got '%s'", expectedTitle, event.Title)
	}

	if event.UID != "test-event-1@example.com" {
		t.Errorf("Expected UID 'test-event-1@example.com', got '%s'", event.UID)
	}

	londonTZ, _ := time.LoadLocation("Europe/London")
	expectedStart := time.Date(2005, 8, 2, 17, 0, 0, 0, londonTZ)
	if !event.Start.Equal(expectedStart) {
		t.Errorf("Expected start time %v, got %v", expectedStart, event.Start)
	}

	expectedEnd := time.Date(2005, 8, 2, 23, 0, 0, 0, londonTZ)
	if !event.Stop.Equal(expectedEnd) {
		t.Errorf("Expected end time %v, got %v", expectedEnd, event.Stop)
	}
}

func TestGetEventsMultiple(t *testing.T) {
	server := serveICSFile("testdata/multiple.ics")
	defer server.Close()

	ctx := context.Background()
	cfg := config.ICal{URL: server.URL}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC)

	events, err := cal.GetEvents(start, end)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) != 3 {
		t.Errorf("Expected 3 events, got %d", len(events))
	}

	expectedTitles := []string{"Test Event 1", "Test Event 2", "Test Event 3"}
	expectedUIDs := []string{"event1@test.com", "event2@test.com", "event3@test.com"}

	for i, event := range events {
		if event.Title != expectedTitles[i] {
			t.Errorf("Event %d: expected title '%s', got '%s'", i, expectedTitles[i], event.Title)
		}
		if event.UID != expectedUIDs[i] {
			t.Errorf("Event %d: expected UID '%s', got '%s'", i, expectedUIDs[i], event.UID)
		}
	}
}

func TestGetEventsDateRange(t *testing.T) {
	server := serveICSFile("testdata/multiple.ics")
	defer server.Close()

	ctx := context.Background()
	cfg := config.ICal{URL: server.URL}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	start := time.Date(2024, 8, 16, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 8, 16, 23, 59, 59, 0, time.UTC)

	events, err := cal.GetEvents(start, end)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event in date range, got %d", len(events))
	}

	if events[0].Title != "Test Event 2" {
		t.Errorf("Expected 'Test Event 2', got '%s'", events[0].Title)
	}
}

func TestGetEventsOutsideDateRange(t *testing.T) {
	server := serveICSFile("testdata/test.ics")
	defer server.Close()

	ctx := context.Background()
	cfg := config.ICal{URL: server.URL}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC)

	events, err := cal.GetEvents(start, end)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("Expected 0 events outside date range, got %d", len(events))
	}
}

func TestGetEventsEmptyCalendar(t *testing.T) {
	server := serveICSFile("testdata/empty.ics")
	defer server.Close()

	ctx := context.Background()
	cfg := config.ICal{URL: server.URL}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC)

	events, err := cal.GetEvents(start, end)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("Expected 0 events for empty calendar, got %d", len(events))
	}
}

func TestGetEventsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer server.Close()

	ctx := context.Background()
	cfg := config.ICal{URL: server.URL}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC)

	events, err := cal.GetEvents(start, end)
	if err != nil && strings.Contains(err.Error(), "failed to parse calendar") {
		return
	}

	if len(events) != 0 {
		t.Errorf("Expected 0 events for HTTP error, got %d", len(events))
	}
}

func TestGetEventsInvalidURL(t *testing.T) {
	ctx := context.Background()
	cfg := config.ICal{URL: "invalid-url"}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC)

	_, err = cal.GetEvents(start, end)
	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}

	if !strings.Contains(err.Error(), "failed to fetch calendar") {
		t.Errorf("Expected fetch error, got %v", err)
	}
}

func TestDeleteAll(t *testing.T) {
	ctx := context.Background()
	cfg := config.ICal{URL: "http://example.com/calendar.ics"}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	err = cal.DeleteAll()
	if err == nil {
		t.Error("Expected error for DeleteAll, got nil")
	}

	expectedMsg := "DeleteAll not implemented for ICS calendar"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestPutEvents(t *testing.T) {
	ctx := context.Background()
	cfg := config.ICal{URL: "http://example.com/calendar.ics"}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	err = cal.PutEvents()
	if err == nil {
		t.Error("Expected error for PutEvents, got nil")
	}

	expectedMsg := "PutEvents not implemented for ICS calendar"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestSyncToDest(t *testing.T) {
	ctx := context.Background()
	cfg := config.ICal{URL: "http://example.com/calendar.ics"}
	cal, err := New(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create calendar: %v", err)
	}

	events := []calendar.Event{{Title: "Test Event"}}
	err = cal.SyncToDest(events)
	if err == nil {
		t.Error("Expected error for SyncToDest, got nil")
	}

	expectedMsg := "SyncToDest not implemented for ICS calendar"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestIsUnknownTZ(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		expected bool
	}{
		{"Known timezone", "Pacific Standard Time", false},
		{"Another known timezone", "GMT Standard Time", false},
		{"Unknown timezone", "Unknown Timezone", true},
		{"Empty string", "", true},
	}

	tzMapping := map[string]string{
		"Pacific Standard Time": "America/Los_Angeles",
		"GMT Standard Time":     "Europe/London",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isUnknownTZ(tzMapping, tt.timezone)
			if result != tt.expected {
				t.Errorf("isUnknownTZ(%s) = %v, expected %v", tt.timezone, result, tt.expected)
			}
		})
	}
}

func TestGetEventsFunction(t *testing.T) {
	server := serveICSFile("testdata/test.ics")
	defer server.Close()

	start := time.Date(2005, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2005, 8, 31, 23, 59, 59, 0, time.UTC)

	events, err := getEvents(server.URL, start, end)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.Title != "Data::ICal release party,other things with slash\\es" {
		t.Errorf("Unexpected event title: %s", event.Title)
	}
}
