package gcal

import (
	"calsync/calendar"
	"calsync/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	gcalendar "google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

var ErrInvalidToken = errors.New("invalid token, delete token.json and try again")
var ErrCalendarNotFound = errors.New("calendar not found")

type Client struct {
	Svc *gcalendar.Service

	http      *http.Client
	cfg       config.Google
	workCalID string
}

func New(ctx context.Context, cfg config.Google, oauthCfg *oauth2.Config) (*Client, error) {
	httpClient := newClient(cfg, oauthCfg)
	svc, err := gcalendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("getting calendar service: %s", err)
	}

	return &Client{
		Svc:       svc,
		http:      httpClient,
		cfg:       cfg,
		workCalID: cfg.Id,
	}, nil
}

func (c *Client) DeleteAll() error {
	return fmt.Errorf("not implemented")
}

func (c *Client) PutEvents() error {
	return fmt.Errorf("not implemented")
}

func (C *Client) GetEvents(_ time.Time, _ time.Time) ([]calendar.Event, error) {
	return nil, fmt.Errorf("not implemented")
}

// NewClient Retrieve a token, saves the token, then returns the generated client.
func newClient(cfg config.Google, oauthCfg *oauth2.Config) *http.Client {
	tokFile := cfg.TokenFile()
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(oauthCfg)
		saveToken(tokFile, tok)
	}
	return oauthCfg.Client(context.Background(), tok)
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
