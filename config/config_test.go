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
		Source: Calendars{
			Mac: &Mac{
				SrcCalBase: SrcCalBase{
					Enabled: true,
				},
				ICalBuddyBinary: "/usr/local/bin/icalBuddy",
				Name:            "Calendar",
			},
			ICal: &ICal{
				SrcCalBase: SrcCalBase{
					Enabled: true,
				},
				URL: "https://ics",
			},
		},
		Target: Calendars{
			Google: &Google{
				Id:          "abcd@group.calendar.google.com",
				Credentials: "credentials.json",
				Token:       "token.json",
			},
		},
	}

	assert.Equal(t, got, expected, "Config should be equal")
}
