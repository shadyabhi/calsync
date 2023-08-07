package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	workCalID = "d70274f2cfca87a126b31d6cb86bb626a436e785372618e3e030faa15b4c58c7@group.calendar.google.com"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
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

func main() {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, calendar.CalendarScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	deleteExisting(srv)
	publishAll(srv)
}

func publishAll(srv *calendar.Service) {
	start := time.Now()
	events, err := getEvents()
	if err != nil {
		panic(err)
	}

	for _, event := range events {
		if err := publishEvent(srv, event); err != nil {
			log.Fatalf("Publishing all events failed: event: %s , %s", event, err)
		}
	}
	log.Printf("Finished adding all events to Google in %s", time.Since(start))

}

func deleteExisting(srv *calendar.Service) {
	log.Printf("Starting deleting of existing events...")
	start := time.Now()
	// Hack: Truncate works with UTC, so we need to include the whole day
	t := time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)
	eventsFromGoogle, err := srv.Events.List(workCalID).ShowDeleted(false).
		SingleEvents(true).TimeMin(t.Format(time.RFC3339)).TimeMax(t.Add(14 * 24 * time.Hour).Format(time.RFC3339)).OrderBy("startTime").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve events from Google: %v", err)
	}
	for _, event := range eventsFromGoogle.Items {
		if err := srv.Events.Delete(workCalID, event.Id).Do(); err != nil {
			log.Fatalf("Cleanup up existing elements failed: %s", err)
		}
	}
	log.Printf("Finished deleting existing events in %s", time.Since(start))
}
