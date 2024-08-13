package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	BoltwoodSource  string `json:"boltwoodSource"`
	PollingInterval string `json:"pollingInterval"`
	WebServerPort   int    `json:"webServerPort"`
	DiscoveryPort   int    `json:"discoveryPort"`
	Timezone        string `json:"timezone"`
}

var config Config

func loadConfig() error {
	// Get the directory of the executable
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	configPath := filepath.Join(filepath.Dir(ex), "config.json")

	// Read the config file
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the JSON data
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	// Validate and process the configuration
	if err := validateConfig(); err != nil {
		return err
	}

	log.Printf("Configuration loaded successfully: %+v", config)
	return nil
}

func validateConfig() error {
	if config.BoltwoodSource == "" {
		return fmt.Errorf("BoltwoodSource is not specified in the config file")
	}

	if config.WebServerPort == 0 {
		return fmt.Errorf("WebServerPort is not specified in the config file")
	}

	// Parse the polling interval
	duration, err := time.ParseDuration(config.PollingInterval)
	if err != nil {
		return fmt.Errorf("invalid PollingInterval in config file: %v", err)
	}
	config.PollingInterval = duration.String()

	// Validate the timezone
	if config.Timezone == "" {
		config.Timezone = "UTC" // Default to UTC if not specified
	} else {
		_, err := time.LoadLocation(config.Timezone)
		if err != nil {
			return fmt.Errorf("invalid Timezone in config file: %v", err)
		}
	}

	return nil
}
