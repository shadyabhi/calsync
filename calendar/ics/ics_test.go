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
	londonTZ, _ := time.LoadLocation("Europe/London")

	tests := []struct {
		name           string
		icsFile        string
		startDate      time.Time
		endDate        time.Time
		expectedCount  int
		expectedError  bool
		errorContains  string
		validateEvents func(t *testing.T, events []calendar.Event)
	}{
		{
			name:          "single event with timezone",
			icsFile:       "testdata/test.ics",
			startDate:     time.Date(2005, 8, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2005, 8, 31, 23, 59, 59, 0, time.UTC),
			expectedCount: 1,
			validateEvents: func(t *testing.T, events []calendar.Event) {
				event := events[0]
				if event.Title != "Data::ICal release party,other things with slash\\es" {
					t.Errorf("Expected title 'Data::ICal release party,other things with slash\\es', got '%s'", event.Title)
				}
				if event.UID != "test-event-1@example.com" {
					t.Errorf("Expected UID 'test-event-1@example.com', got '%s'", event.UID)
				}
				expectedStart := time.Date(2005, 8, 2, 17, 0, 0, 0, londonTZ)
				if !event.Start.Equal(expectedStart) {
					t.Errorf("Expected start time %v, got %v", expectedStart, event.Start)
				}
				expectedEnd := time.Date(2005, 8, 2, 23, 0, 0, 0, londonTZ)
				if !event.Stop.Equal(expectedEnd) {
					t.Errorf("Expected end time %v, got %v", expectedEnd, event.Stop)
				}
			},
		},
		{
			name:          "multiple events",
			icsFile:       "testdata/multiple.ics",
			startDate:     time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC),
			expectedCount: 3,
			validateEvents: func(t *testing.T, events []calendar.Event) {
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
			},
		},
		{
			name:          "filtered date range",
			icsFile:       "testdata/multiple.ics",
			startDate:     time.Date(2024, 8, 16, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2024, 8, 16, 23, 59, 59, 0, time.UTC),
			expectedCount: 1,
			validateEvents: func(t *testing.T, events []calendar.Event) {
				if events[0].Title != "Test Event 2" {
					t.Errorf("Expected 'Test Event 2', got '%s'", events[0].Title)
				}
			},
		},
		{
			name:          "events outside date range",
			icsFile:       "testdata/test.ics",
			startDate:     time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC),
			expectedCount: 0,
		},
		{
			name:          "empty calendar",
			icsFile:       "testdata/empty.ics",
			startDate:     time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC),
			expectedCount: 0,
		},
		{
			name:          "non-existent file",
			icsFile:       "testdata/nonexistent.ics",
			startDate:     time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC),
			endDate:       time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC),
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := serveICSFile(tt.icsFile)
			defer server.Close()

			ctx := context.Background()
			cfg := config.ICal{URL: server.URL}
			cal, err := New(ctx, cfg)
			if err != nil {
				t.Fatalf("Failed to create calendar: %v", err)
			}

			events, err := cal.GetEvents(tt.startDate, tt.endDate)
			if tt.expectedError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got '%v'", tt.errorContains, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(events) != tt.expectedCount {
				t.Errorf("Expected %d events, got %d", tt.expectedCount, len(events))
			}

			if tt.validateEvents != nil {
				tt.validateEvents(t, events)
			}
		})
	}
}

func TestGetEventsErrors(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		expectedError string
	}{
		{
			name:          "invalid URL",
			url:           "invalid-url",
			expectedError: "failed to fetch calendar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			cfg := config.ICal{URL: tt.url}
			cal, err := New(ctx, cfg)
			if err != nil {
				t.Fatalf("Failed to create calendar: %v", err)
			}

			start := time.Date(2024, 8, 1, 0, 0, 0, 0, time.UTC)
			end := time.Date(2024, 8, 31, 23, 59, 59, 0, time.UTC)

			_, err = cal.GetEvents(start, end)
			if err == nil {
				t.Error("Expected error, got nil")
			} else if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got '%v'", tt.expectedError, err)
			}
		})
	}
}
