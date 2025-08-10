package config

import (
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	got, err := GetConfig("testdata/config.toml")
	if err != nil {
		slog.Error("Failed to get config", "error", err)
		os.Exit(1)
	}
	expected := &Config{
		Source: Calendars{
			Mac: &Mac{
				SrcCalBase: SrcCalBase{
					Enabled: false,
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
				SrcCalBase: SrcCalBase{
					Enabled: true,
				},
				Id:          "abcd@group.calendar.google.com",
				Credentials: "credentials.json",
				Token:       "token.json",
			},
		},
		Sync: Sync{
			Days: 14,
		},
	}

	assert.Equal(t, got, expected, "Config should be equal")
}
