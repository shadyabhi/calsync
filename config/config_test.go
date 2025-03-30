package config

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	got, err := GetConfig("testdata/config.toml")
	if err != nil {
		log.Fatalf("Failed to get config: %s", err)
	}
	expected := &Config{
		Secrets: struct {
			Credentials string
			Token       string
		}{
			Credentials: "credentials.json",
			Token:       "token.json",
		},
		Source: Calendars{
			Mac: &Mac{
				SrcCalBase: SrcCalBase{
					Enabled: true,
					Days:    7,
				},
				ICalBuddyBinary: "/usr/local/bin/icalBuddy",
				Name:            "Calendar",
			},
			ICal: &ICal{
				SrcCalBase: SrcCalBase{
					Enabled: true,
					Days:    7,
				},
				URL: "https://ics",
			},
		},
		Target: Calendars{
			Google: &Google{
				Id: "abcd@group.calendar.google.com",
			},
		},
	}

	assert.Equal(t, got, expected, "Config should be equal")
}
