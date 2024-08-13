package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func main() {
	// Load configuration
	if err := loadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Start weather data polling and register the driver with alpaca
	go pollWeatherData()
	go handleAlpacaDiscovery()

	// Setup and start web server
	router := mux.NewRouter()
	loggedHandler := setupRoutes(router)

	// Create a new http.Server with the logged handler
	server := &http.Server{
		Addr:    fmt.Sprintf("127.0.0.1:%d", config.WebServerPort),
		Handler: loggedHandler,
	}

	log.Printf("Starting web server on http://localhost:%d", config.WebServerPort)
	log.Fatal(server.ListenAndServe())
}

func parseDaylightCondition(val int) string {
	switch val {
	case 0:
		return "Unknown"
	case 1:
		return "Dark"
	case 2:
		return "Light"
	case 3:
		return "Very Light"
	default:
		return "Unknown"
	}
}

func parseCloudCondition(val int) string {
	switch val {
	case 1:
		return "Clear"
	case 2:
		return "Light Clouds"
	case 3:
		return "Very Cloudy"
	default:
		return "Unknown"
	}
}

func parseWindCondition(val int) string {
	switch val {
	case 1:
		return "Calm"
	case 2:
		return "Windy"
	case 3:
		return "Very Windy"
	default:
		return "Unknown"
	}
}

func parseRainCondition(val int) string {
	switch val {
	case 1:
		return "Dry"
	case 2:
		return "Damp"
	case 3:
		return "Rain"
	default:
		return "Unknown"
	}
}

func parseDarknessCondition(val int) string {
	switch val {
	case 1:
		return "Dark"
	case 2:
		return "Dim"
	case 3:
		return "Daylight"
	default:
		return "Unknown"
	}
}

func parseAlertStatus(val int) string {
	switch val {
	case 0:
		return "No Alert"
	case 1:
		return "Alert"
	default:
		return "Unknown"
	}
}
