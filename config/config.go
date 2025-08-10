package config

import (
	"calsync/calendar"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Version string

	Source Calendars
	Target Calendars

	Sync Sync
}

type Calendars struct {
	Mac    *Mac
	ICal   *ICal
	Google *Google
}
type SrcCalBase struct {
	Enabled bool
	Cal     calendar.Calendar
}
type Mac struct {
	SrcCalBase

	ICalBuddyBinary string
	Name            string
}

type ICal struct {
	SrcCalBase

	URL string
}

type Google struct {
	SrcCalBase

	Id          string
	Credentials string
	Token       string
}

type Sync struct {
	Days int
}

func (g Google) TokenFile() string {
	return filepath.Join(
		os.Getenv("HOME"),
		"/.config/calsync/",
		"token.json",
	)
}

func (c Config) CredentialsFile() string {
	return filepath.Join(
		os.Getenv("HOME"),
		"/.config/calsync/",
		"credentials.json",
	)
}

func (c Config) ConfigFile() string {
	return filepath.Join(
		os.Getenv("HOME"),
		"/.config/calsync/",
		"config.toml",
	)
}

func GetConfig(location string) (*Config, error) {
	if _, err := os.Stat(location); os.IsNotExist(err) {
		slog.Error("Config file not found", "location", location)
		os.Exit(1)
	}

	config := &Config{}
	_, err := toml.DecodeFile(location, config)
	if err != nil {
		return config, fmt.Errorf("Failed to decode config file: %s", err)
	}

	return config, nil
}
