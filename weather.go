package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type WeatherData struct {
	Date                time.Time `json:"date"`
	TemperatureScale    string    `json:"temperatureScale"`
	WindSpeedScale      string    `json:"windSpeedScale"`
	SkyTemperature      float64   `json:"skyTemperature"`
	AmbientTemperature  float64   `json:"ambientTemperature"`
	SensorTemperature   float64   `json:"sensorTemperature"`
	WindSpeed           float64   `json:"windSpeed"`
	Humidity            float64   `json:"humidity"`
	DewPoint            float64   `json:"dewPoint"`
	DewHeaterPercentage float64   `json:"dewHeaterPercentage"`
	RainFlag            int       `json:"rainFlag"`
	WetFlag             int       `json:"wetFlag"`
	CloudCondition      string    `json:"cloudCondition"`
	WindCondition       string    `json:"windCondition"`
	RainCondition       string    `json:"rainCondition"`
	DarknessCondition   string    `json:"darknessCondition"`
	AlertStatus         string    `json:"alertStatus"`
}

var weatherData WeatherData

func pollWeatherData() {
	interval, _ := time.ParseDuration(config.PollingInterval)
	for {
		data, err := readBoltwoodData(config.BoltwoodSource)
		if err != nil {
			log.Printf("Error reading Boltwood data: %v", err)
		} else {
			parseAndUpdateWeatherData(data)
		}
		time.Sleep(interval)
	}
}

func readBoltwoodData(source string) ([]byte, error) {
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return readFromHTTP(source)
	}
	return readFromFile(source)
}

func readFromHTTP(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func readFromFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func parseAndUpdateWeatherData(data []byte) {
	lines := strings.Split(string(data), "\n")
	if len(lines) != 1 {
		log.Printf("Invalid Boltwood data format")
		return
	}

	fields := strings.Fields(lines[0])
	if len(fields) < 21 {
		log.Printf("Insufficient fields in Boltwood data: got %d, expected 21", len(fields))
		return
	}

	var err error
	var newWeatherData WeatherData

	// Parse date and time (fields 0 and 1)
	dateStr := fields[0] + " " + fields[1]

	// Load the configured timezone
	loc, err := time.LoadLocation(config.Timezone)
	if err != nil {
		log.Printf("Error loading timezone %s: %v", config.Timezone, err)
		return
	}

	// Parse the date string in the configured timezone
	newWeatherData.Date, err = time.ParseInLocation("2006-01-02 15:04:05.00", dateStr, loc)
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		return
	}

	// Convert the time to UTC for storage
	newWeatherData.Date = newWeatherData.Date.UTC()

	newWeatherData.TemperatureScale = fields[2]
	newWeatherData.WindSpeedScale = fields[3]
	newWeatherData.SkyTemperature, _ = strconv.ParseFloat(fields[4], 64)
	newWeatherData.AmbientTemperature, _ = strconv.ParseFloat(fields[5], 64)
	newWeatherData.SensorTemperature, _ = strconv.ParseFloat(fields[6], 64)
	newWeatherData.WindSpeed, _ = strconv.ParseFloat(fields[7], 64)
	newWeatherData.Humidity, _ = strconv.ParseFloat(fields[8], 64)
	newWeatherData.DewPoint, _ = strconv.ParseFloat(fields[9], 64)
	newWeatherData.DewHeaterPercentage, _ = strconv.ParseFloat(fields[10], 64)
	newWeatherData.RainFlag, _ = strconv.Atoi(fields[11])
	newWeatherData.WetFlag, _ = strconv.Atoi(fields[12])
	// These fields are garbage so ignore them
	//newWeatherData.ElapsedTime, _ = strconv.ParseFloat(fields[13], 64)
	//newWeatherData.ElapsedDays, _ = strconv.ParseFloat(fields[14], 64)

	cloudVal, _ := strconv.Atoi(fields[15])
	newWeatherData.CloudCondition = parseCloudCondition(cloudVal)

	windVal, _ := strconv.Atoi(fields[16])
	newWeatherData.WindCondition = parseWindCondition(windVal)

	rainVal, _ := strconv.Atoi(fields[17])
	newWeatherData.RainCondition = parseRainCondition(rainVal)

	darknessVal, _ := strconv.Atoi(fields[18])
	newWeatherData.DarknessCondition = parseDarknessCondition(darknessVal)

	alertVal, _ := strconv.Atoi(fields[20])
	newWeatherData.AlertStatus = parseAlertStatus(alertVal)

	// Update the global weatherData
	weatherData = newWeatherData
	log.Printf("Weather data updated: %+v", weatherData)
}
