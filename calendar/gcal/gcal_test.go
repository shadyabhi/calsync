package gcal

import (
	"calsync/calendar"
	"calsync/config"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/oauth2"
	googlecalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

func TestSyncToDest(t *testing.T) {
	tests := []struct {
		name           string
		existingEvents []*googlecalendar.Event
		localEvents    []calendar.Event
		wantCreated    int
		wantDeleted    int
		wantDeletedIDs []string
	}{
		{
			name:           "sync new events",
			existingEvents: []*googlecalendar.Event{},
			localEvents: []calendar.Event{
				{
					Title: "Test Event 1",
					Notes: "Test Notes 1",
					Start: time.Now().Add(1 * time.Hour),
					Stop:  time.Now().Add(2 * time.Hour),
					UID:   "uid1",
				},
				{
					Title: "Test Event 2",
					Notes: "Test Notes 2",
					Start: time.Now().Add(3 * time.Hour),
					Stop:  time.Now().Add(4 * time.Hour),
					UID:   "uid2",
				},
			},
			wantCreated:    2,
			wantDeleted:    0,
			wantDeletedIDs: []string{},
		},
		{
			name: "delete stale calsync events",
			existingEvents: []*googlecalendar.Event{
				{
					Id:      "stale-event-1",
					Summary: "Stale Event",
					Start: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					},
					End: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					},
					Source: &googlecalendar.EventSource{
						Title: EventSourceTitle,
					},
				},
			},
			localEvents: []calendar.Event{
				{
					Title: "New Event",
					Notes: "New Notes",
					Start: time.Now().Add(1 * time.Hour),
					Stop:  time.Now().Add(2 * time.Hour),
					UID:   "uid1",
				},
			},
			wantCreated:    1,
			wantDeleted:    1,
			wantDeletedIDs: []string{"stale-event-1"},
		},
		{
			name: "skip non-calsync events",
			existingEvents: []*googlecalendar.Event{
				{
					Id:      "manual1",
					Summary: "Manual Event",
					Start: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					},
					End: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					},
					Source: &googlecalendar.EventSource{
						Title: "Other Source",
					},
				},
			},
			localEvents: []calendar.Event{
				{
					Title: "New Event",
					Notes: "New Notes",
					Start: time.Now().Add(1 * time.Hour),
					Stop:  time.Now().Add(2 * time.Hour),
					UID:   "uid1",
				},
			},
			wantCreated:    1,
			wantDeleted:    0,
			wantDeletedIDs: []string{},
		},
		{
			name: "keep matching events",
			existingEvents: []*googlecalendar.Event{
				{
					Id:          "existing1",
					Summary:     "Matching Event",
					Description: "Matching Notes",
					Start: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					},
					End: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					},
					Source: &googlecalendar.EventSource{
						Title: EventSourceTitle,
						Url:   "https://github.com/shadyabhi/calsync",
					},
					ExtendedProperties: &googlecalendar.EventExtendedProperties{
						Private: map[string]string{
							"uid": "uid1",
						},
					},
				},
			},
			localEvents: []calendar.Event{
				{
					Title: "Matching Event",
					Notes: "Matching Notes",
					Start: time.Now().Add(1 * time.Hour),
					Stop:  time.Now().Add(2 * time.Hour),
					UID:   "uid1",
				},
			},
			wantCreated:    0,
			wantDeleted:    0,
			wantDeletedIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := newMockServer(t)
			defer mockServer.Close()

			// Add existing events to mock server
			for _, event := range tt.existingEvents {
				mockServer.addEvent(event)
			}

			testConfig := newTestClientConfig(t, mockServer)
			client := newTestClient(testConfig.Service, testConfig.HTTPClient, testConfig.Config)

			// Run sync
			err := client.SyncToDest(tt.localEvents)
			if err != nil {
				t.Fatalf("SyncToDest failed: %v", err)
			}

			// Verify created events
			createdCount := mockServer.CreatedCount
			if createdCount != tt.wantCreated {
				t.Errorf("Created events: got %d, want %d", createdCount, tt.wantCreated)
			}

			// Verify deleted events
			deletedCount := len(mockServer.DeletedIDs)
			if deletedCount != tt.wantDeleted {
				t.Errorf("Deleted events: got %d, want %d", deletedCount, tt.wantDeleted)
			}

			// Verify specific deleted IDs
			for _, wantID := range tt.wantDeletedIDs {
				found := false
				for _, deletedID := range mockServer.DeletedIDs {
					if deletedID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected event %s to be deleted, but it wasn't", wantID)
				}
			}

			// Verify created events have correct source and properties
			for _, event := range mockServer.Events {
				if event.Source == nil || event.Source.Title != EventSourceTitle {
					continue // Skip pre-existing test events
				}

				// Verify source is set correctly
				if event.Source.Title != EventSourceTitle {
					t.Errorf("Event source title: got %s, want %s", event.Source.Title, EventSourceTitle)
				}

				// Verify source URL is set
				if event.Source.Url != "https://github.com/shadyabhi/calsync" {
					t.Errorf("Event source URL: got %s, want https://github.com/shadyabhi/calsync", event.Source.Url)
				}

				// Verify extended properties are set
				if event.ExtendedProperties == nil || event.ExtendedProperties.Private == nil {
					t.Error("Event is missing extended properties")
				}
			}
		})
	}
}

func TestDeleteAllInRange(t *testing.T) {
	tests := []struct {
		name           string
		existingEvents []*googlecalendar.Event
		start          time.Time
		end            time.Time
		wantDeleted    []string
	}{
		{
			name: "delete only calsync events in range",
			existingEvents: []*googlecalendar.Event{
				{
					Id:      "calsync1",
					Summary: "CalSync Event",
					Start: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					},
					End: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					},
					Source: &googlecalendar.EventSource{
						Title: EventSourceTitle,
					},
				},
				{
					Id:      "manual1",
					Summary: "Manual Event",
					Start: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					},
					End: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					},
				},
			},
			start:       time.Now(),
			end:         time.Now().Add(24 * time.Hour),
			wantDeleted: []string{"calsync1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := newMockServer(t)
			defer mockServer.Close()

			// Add existing events
			for _, event := range tt.existingEvents {
				mockServer.addEvent(event)
			}

			testConfig := newTestClientConfig(t, mockServer)
			client := newTestClient(testConfig.Service, testConfig.HTTPClient, testConfig.Config)

			// Run delete
			err := client.DeleteAllInRange(tt.start, tt.end)
			if err != nil {
				t.Fatalf("DeleteAllInRange failed: %v", err)
			}

			// Verify deleted events
			if len(mockServer.DeletedIDs) != len(tt.wantDeleted) {
				t.Errorf("Deleted count: got %d, want %d", len(mockServer.DeletedIDs), len(tt.wantDeleted))
			}

			for _, wantID := range tt.wantDeleted {
				found := false
				for _, deletedID := range mockServer.DeletedIDs {
					if deletedID == wantID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected event %s to be deleted", wantID)
				}
			}
		})
	}
}

func TestGetAllGCalEvents(t *testing.T) {
	tests := []struct {
		name           string
		existingEvents []*googlecalendar.Event
		start          time.Time
		end            time.Time
		wantCount      int
	}{
		{
			name: "get events in range",
			existingEvents: []*googlecalendar.Event{
				{
					Id:      "event1",
					Summary: "Event 1",
					Start: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
					},
					End: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(2 * time.Hour).Format(time.RFC3339),
					},
				},
				{
					Id:      "event2",
					Summary: "Event 2",
					Start: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(25 * time.Hour).Format(time.RFC3339),
					},
					End: &googlecalendar.EventDateTime{
						DateTime: time.Now().Add(26 * time.Hour).Format(time.RFC3339),
					},
				},
			},
			start:     time.Now(),
			end:       time.Now().Add(24 * time.Hour),
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := newMockServer(t)
			defer mockServer.Close()

			// Add existing events
			for _, event := range tt.existingEvents {
				mockServer.addEvent(event)
			}

			testConfig := newTestClientConfig(t, mockServer)
			client := newTestClient(testConfig.Service, testConfig.HTTPClient, testConfig.Config)

			// Get events
			events, err := client.GetAllGCalEvents(tt.start, tt.end)
			if err != nil {
				t.Fatalf("GetAllGCalEvents failed: %v", err)
			}

			if len(events) != tt.wantCount {
				t.Errorf("Event count: got %d, want %d", len(events), tt.wantCount)
			}
		})
	}
}

func TestPublishAllEvents(t *testing.T) {
	mockServer := newMockServer(t)
	defer mockServer.Close()

	testConfig := newTestClientConfig(t, mockServer)
	client := newTestClient(testConfig.Service, testConfig.HTTPClient, testConfig.Config)

	events := []calendar.Event{
		{
			Title: "Test Event 1",
			Notes: "Notes 1",
			Start: time.Now().Add(1 * time.Hour),
			Stop:  time.Now().Add(2 * time.Hour),
			UID:   "uid1",
		},
		{
			Title: "Test Event 2",
			Notes: "Notes 2",
			Start: time.Now().Add(3 * time.Hour),
			Stop:  time.Now().Add(4 * time.Hour),
			UID:   "uid2",
		},
	}

	err := client.PublishAllEvents(events)
	if err != nil {
		t.Fatalf("PublishAllEvents failed: %v", err)
	}

	if mockServer.CreatedCount != 2 {
		t.Errorf("Created events: got %d, want %d", mockServer.CreatedCount, 2)
	}

	// Verify event details
	for i, mockEvent := range mockServer.Events {
		if mockEvent.Summary != events[i].Title {
			t.Errorf("Event title: got %s, want %s", mockEvent.Summary, events[i].Title)
		}
		if mockEvent.Description != events[i].Notes {
			t.Errorf("Event notes: got %s, want %s", mockEvent.Description, events[i].Notes)
		}
	}
}

// TestWithOAuth2Mock demonstrates how to test with OAuth2 mocking
func TestWithOAuth2Mock(t *testing.T) {
	mockServer := newMockServer(t)
	defer mockServer.Close()

	// Create mock OAuth2 config
	oauthConfig := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  mockServer.URL + "/auth",
			TokenURL: mockServer.URL + "/token",
		},
		Scopes: []string{googlecalendar.CalendarScope},
	}

	// Create a test token
	token := &oauth2.Token{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour),
	}

	// Create HTTP client with token
	httpClient := oauthConfig.Client(context.Background(), token)

	// Create calendar service
	ctx := context.Background()
	svc, err := googlecalendar.NewService(ctx,
		option.WithHTTPClient(httpClient),
		option.WithEndpoint(mockServer.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	cfg := config.Google{Id: "test-calendar"}
	client := newTestClient(svc, httpClient, cfg)

	// Test a simple operation
	events, err := client.GetAllGCalEvents(time.Now(), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Fatalf("GetAllGCalEvents failed: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("Expected 0 events, got %d", len(events))
	}
}

// Test helper types and functions

// newTestClient creates a Client for testing with custom service and HTTP client
func newTestClient(svc *googlecalendar.Service, httpClient *http.Client, cfg config.Google) *Client {
	return &Client{
		Svc:       svc,
		http:      httpClient,
		cfg:       cfg,
		workCalID: cfg.Id,
	}
}

type testClientConfig struct {
	Service    *googlecalendar.Service
	HTTPClient *http.Client
	Config     config.Google
}

func newTestClientConfig(t *testing.T, mockServer *mockServer) *testClientConfig {
	t.Helper()

	// Create a custom HTTP client that doesn't require authentication
	httpClient := &http.Client{}

	// Create the calendar service with custom endpoint
	ctx := context.Background()
	svc, err := googlecalendar.NewService(ctx,
		option.WithHTTPClient(httpClient),
		option.WithEndpoint(mockServer.URL),
	)
	if err != nil {
		t.Fatalf("Failed to create calendar service: %v", err)
	}

	return &testClientConfig{
		Service:    svc,
		HTTPClient: httpClient,
		Config:     config.Google{Id: "test-calendar"},
	}
}

type mockServer struct {
	*httptest.Server
	Events       []*googlecalendar.Event
	DeletedIDs   []string
	CreatedCount int
	t            *testing.T
}

func newMockServer(t *testing.T) *mockServer {
	m := &mockServer{
		Events:     []*googlecalendar.Event{},
		DeletedIDs: []string{},
		t:          t,
	}

	mux := http.NewServeMux()

	// Mock events list endpoint
	mux.HandleFunc("/calendars/test-calendar/events", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			m.handleListEvents(w, r)
		case "POST":
			m.handleCreateEvent(w, r)
		}
	})

	// Mock individual event operations
	mux.HandleFunc("/calendars/test-calendar/events/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "DELETE" {
			m.handleDeleteEvent(w, r)
		}
	})

	// Mock calendar list endpoint
	mux.HandleFunc("/users/me/calendarList", func(w http.ResponseWriter, r *http.Request) {
		m.handleCalendarList(w, r)
	})

	m.Server = httptest.NewServer(mux)
	return m
}

func (m *mockServer) handleListEvents(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	timeMin := query.Get("timeMin")
	timeMax := query.Get("timeMax")

	var filteredEvents []*googlecalendar.Event

	for _, event := range m.Events {
		if event.Start == nil || event.Start.DateTime == "" {
			continue
		}

		eventTime, _ := time.Parse(time.RFC3339, event.Start.DateTime)
		minTime, _ := time.Parse(time.RFC3339, timeMin)
		maxTime, _ := time.Parse(time.RFC3339, timeMax)

		if !eventTime.Before(minTime) && !eventTime.After(maxTime) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	response := &googlecalendar.Events{
		Items: filteredEvents,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		m.t.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (m *mockServer) handleCreateEvent(w http.ResponseWriter, r *http.Request) {
	var event googlecalendar.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	event.Id = fmt.Sprintf("event-%d", m.CreatedCount)
	m.CreatedCount++
	m.Events = append(m.Events, &event)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(&event); err != nil {
		m.t.Errorf("Failed to encode event: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (m *mockServer) handleDeleteEvent(w http.ResponseWriter, r *http.Request) {
	// Extract event ID from URL path
	path := r.URL.Path
	eventID := path[len("/calendars/test-calendar/events/"):]

	m.DeletedIDs = append(m.DeletedIDs, eventID)

	// Remove from events list
	for i, event := range m.Events {
		if event.Id == eventID {
			m.Events = append(m.Events[:i], m.Events[i+1:]...)
			break
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (m *mockServer) handleCalendarList(w http.ResponseWriter, _ *http.Request) {
	response := &googlecalendar.CalendarList{
		Items: []*googlecalendar.CalendarListEntry{
			{
				Id:      "test-calendar",
				Summary: "Test Calendar",
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		m.t.Errorf("Failed to encode calendar list response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (m *mockServer) addEvent(event *googlecalendar.Event) {
	m.Events = append(m.Events, event)
}
