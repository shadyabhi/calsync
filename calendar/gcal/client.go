package gcal

import (
	"calsync/config"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type Client struct {
	Svc *calendar.Service

	http      *http.Client
	workCalID string
}

func New(ctx context.Context, cfg *config.Config, oauthCfg *oauth2.Config) (*Client, error) {
	httpClient := newClient(cfg, oauthCfg)
	svc, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("getting calendar service: %s", err)
	}

	return &Client{
		Svc:       svc,
		http:      httpClient,
		workCalID: cfg.Google.Id,
	}, nil
}

// NewClient Retrieve a token, saves the token, then returns the generated client.
func newClient(cfg *config.Config, oauthCfg *oauth2.Config) *http.Client {
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
