package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

// loggingMiddleware is a function that logs the incoming HTTP request
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log the request details
		log.Printf(
			"%s %s %s %s",
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
			time.Since(start),
		)
	})
}

func setupRoutes(router *mux.Router) http.Handler {
	// Wrap the router with the logging middleware
	loggedRouter := loggingMiddleware(router)

	// Existing routes
	router.HandleFunc("/", handleHome).Methods("GET")
	router.HandleFunc("/api/weather", handleWeatherAPI).Methods("GET")
	router.HandleFunc("/status", handleStatus).Methods("GET")
	router.HandleFunc("/weather", handleWeather).Methods("GET")

	// Alpaca management API endpoints
	router.HandleFunc("/management/apiversions", handleAPIVersions).Methods("GET")
	router.HandleFunc("/management/v1/configureddevices", handleConfiguredDevices).Methods("GET")
	router.HandleFunc("/management/v1/description", handleManagementDescription).Methods("GET")

	// Alpaca device API endpoints
	router.HandleFunc("/api/v1/observingconditions/0/connected", handleConnected).Methods("GET", "PUT")
	router.HandleFunc("/api/v1/observingconditions/0/description", handleDescription).Methods("GET")
	router.HandleFunc("/api/v1/observingconditions/0/driverinfo", handleDriverInfo).Methods("GET")
	router.HandleFunc("/api/v1/observingconditions/0/driverversion", handleDriverVersion).Methods("GET")
	router.HandleFunc("/api/v1/observingconditions/0/name", handleName).Methods("GET")
	router.HandleFunc("/api/v1/observingconditions/0/supportedactions", handleSupportedActions).Methods("GET")
	router.HandleFunc("/api/v1/observingconditions/0/interfaceversion", handleInterfaceVersion).Methods("GET")

	// Existing device-specific endpoints
	router.HandleFunc("/api/v1/observingconditions/0/temperature", handleTemperature).Methods("GET")
	router.HandleFunc("/api/v1/observingconditions/0/humidity", handleHumidity).Methods("GET")
	router.HandleFunc("/api/v1/observingconditions/0/dewpoint", handleDewPoint).Methods("GET")
	router.HandleFunc("/api/v1/observingconditions/0/windspeed", handleWindSpeed).Methods("GET")
	router.HandleFunc("/api/v1/observingconditions/0/windspeed", handleWindSpeed).Methods("GET")

	// Return the logged router instead of the original router
	return loggedRouter
}
