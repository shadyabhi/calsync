package gcal

import (
	"calsync/event"
	"calsync/maccalendar"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	WorkCalID = "d70274f2cfca87a126b31d6cb86bb626a436e785372618e3e030faa15b4c58c7@group.calendar.google.com"
)

type Client struct {
	http *http.Client
	Svc  *calendar.Service
}

func New(ctx context.Context, config *oauth2.Config) (*Client, error) {
	httpClient := newClient(config)
	svc, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("getting calendar service: %s", err)
	}

	return &Client{
		http: httpClient,
		Svc:  svc,
	}, nil
}

func (c *Client) PublishAll() {
	start := time.Now()
	events, err := maccalendar.GetEvents()
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		if err := c.PublishEvent(event); err != nil {
			log.Fatalf("Publishing all events failed: event: %s , %s", event, err)
		}
	}
	log.Printf("Finished adding all events to Google in %s", time.Since(start))
}

func (c *Client) PublishEvent(event event.Event) error {
	calEntry := &calendar.Event{
		Summary: event.Title,
		Start: &calendar.EventDateTime{
			DateTime: event.Start.Format(time.RFC3339),
		},
		End: &calendar.EventDateTime{
			DateTime: event.Stop.Format(time.RFC3339),
		},
		Description: fmt.Sprintf("uid=%s", event.UID),
	}

	calEntry, err := c.Svc.Events.Insert(WorkCalID, calEntry).Do()
	if err != nil {
		return err
	}

	log.Printf("Event created: %s, %s, %s\n", calEntry.Summary, calEntry.Start.DateTime, calEntry.End.DateTime)

	return nil
}

func (c *Client) DeleteAll() {
	start := time.Now()
	log.Printf("Starting deletion of existing events...")

	// Hack: Truncate works with UTC, so we need to include the whole day
	t := time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)
	eventsFromGoogle, err := c.Svc.Events.List(WorkCalID).ShowDeleted(false).
		SingleEvents(true).TimeMin(t.Format(time.RFC3339)).TimeMax(t.Add(14 * 24 * time.Hour).Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve events from Google: %v", err)
	}

	for _, event := range eventsFromGoogle.Items {
		if err := c.Svc.Events.Delete(WorkCalID, event.Id).Do(); err != nil {
			log.Fatalf("Cleanup up existing elements failed: %s", err)
		}
	}

	log.Printf("Finished deleting existing events in %s", time.Since(start))

}

// NewClient Retrieve a token, saves the token, then returns the generated client.
func newClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	_ = json.NewEncoder(f).Encode(token)
}
