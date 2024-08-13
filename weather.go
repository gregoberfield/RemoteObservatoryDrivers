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

	// Temperature scale (field 2)
	tempScale := fields[2]

	// Wind speed scale (field 3)
	windScale := fields[3]

	// Parse and convert temperatures
	newWeatherData.SkyTemperature = convertTemperature(fields[4], tempScale)
	newWeatherData.AmbientTemperature = convertTemperature(fields[5], tempScale)
	newWeatherData.SensorTemperature = convertTemperature(fields[6], tempScale)

	// Parse and convert wind speed
	newWeatherData.WindSpeed = convertWindSpeed(fields[7], windScale)

	newWeatherData.Humidity, _ = strconv.ParseFloat(fields[8], 64)
	newWeatherData.DewPoint = convertTemperature(fields[9], tempScale)
	newWeatherData.DewHeaterPercentage, _ = strconv.ParseFloat(fields[10], 64)
	newWeatherData.RainFlag, _ = strconv.Atoi(fields[11])
	newWeatherData.WetFlag, _ = strconv.Atoi(fields[12])

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

	// Set standard units
	newWeatherData.TemperatureScale = "C"
	newWeatherData.WindSpeedScale = "m/s"

	// Update the global weatherData
	weatherData = newWeatherData
	log.Printf("Weather data updated: %+v", weatherData)
}

func convertTemperature(tempStr, scale string) float64 {
	temp, err := strconv.ParseFloat(tempStr, 64)
	if err != nil {
		log.Printf("Error parsing temperature: %v", err)
		return 0
	}

	if strings.ToUpper(scale) == "F" {
		// Convert Fahrenheit to Celsius
		return (temp - 32) * 5 / 9
	}

	return temp // Already in Celsius
}

func convertWindSpeed(speedStr, scale string) float64 {
	speed, err := strconv.ParseFloat(speedStr, 64)
	if err != nil {
		log.Printf("Error parsing wind speed: %v", err)
		return 0
	}

	if strings.ToUpper(scale) == "M" {
		// Convert mph to m/s
		return speed * 0.44704
	}

	return speed // Assume it's already in m/s if not mph
}
