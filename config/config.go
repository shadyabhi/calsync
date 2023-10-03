package config

import (
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Secrets struct {
		Credentials string
		Token       string
	}

	Mac struct {
		Name string
		Days int
	}

	Google struct {
		Id string
	}
}

func GetConfig(location string) (*Config, error) {
	if _, err := os.Stat(location); os.IsNotExist(err) {
		log.Fatalf("Config file not found at %s", location)
	}

	config := &Config{}
	_, err := toml.DecodeFile(location, config)
	if err != nil {
		return config, fmt.Errorf("Failed to decode config file: %s", err)
	}

	return config, nil
}
