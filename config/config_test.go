package config

import (
	"log"
	"reflect"
	"testing"
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
		Mac: struct {
			ICalBuddyBinary string
			Name            string
			Days            int
		}{
			ICalBuddyBinary: "/usr/local/bin/icalBuddy",
			Name:            "Calendar",
			Days:            7,
		},
		Google: struct {
			Id string
		}{
			Id: "abcd@group.calendar.google.com",
		},
	}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Unexpected parsed config %#v", got)
	}
}
