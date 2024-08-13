package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// Global variable to store the connected state
var (
	isConnected bool
	connMutex   sync.Mutex
)

func handleHome(w http.ResponseWriter, r *http.Request) {
	tmpl := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Boltwood II Weather Data</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; padding: 20px; }
        h1 { color: #333; }
        #weather-data { background: #f4f4f4; padding: 20px; border-radius: 5px; }
        .data-item { margin-bottom: 10px; }
    </style>
</head>
<body>
    <h1>Current Weather Conditions</h1>
    <div id="weather-data">Loading...</div>

    <script>
        const pollingInterval = {{.PollingInterval}};
        
        function updateWeatherData() {
            fetch('/api/weather')
                .then(response => response.json())
                .then(data => {
                    const weatherDiv = document.getElementById('weather-data');
                    weatherDiv.innerHTML = generateWeatherHTML(data);
                })
                .catch(error => console.error('Error fetching weather data:', error));
        }

        function generateWeatherHTML(data) {
            return ` + "`" + `
                <div class="data-item">Date: ${new Date(data.date).toLocaleString()}</div>
                <div class="data-item">Sky Temperature: ${data.skyTemperature.toFixed(1)}${data.temperatureScale}</div>
                <div class="data-item">Ambient Temperature: ${data.ambientTemperature.toFixed(1)}${data.temperatureScale}</div>
                <div class="data-item">Sensor Temperature: ${data.sensorTemperature.toFixed(1)}${data.temperatureScale}</div>
                <div class="data-item">Wind Speed: ${data.windSpeed.toFixed(1)} ${data.windSpeedScale}</div>
                <div class="data-item">Humidity: ${data.humidity.toFixed(1)}%</div>
                <div class="data-item">Dew Point: ${data.dewPoint.toFixed(1)}${data.temperatureScale}</div>
                <div class="data-item">Dew Heater: ${data.dewHeaterPercentage.toFixed(1)}%</div>
                <div class="data-item">Rain Flag: ${data.rainFlag}</div>
                <div class="data-item">Wet Flag: ${data.wetFlag}</div>
                <div class="data-item">Cloud Condition: ${data.cloudCondition}</div>
                <div class="data-item">Wind Condition: ${data.windCondition}</div>
                <div class="data-item">Rain Condition: ${data.rainCondition}</div>
                <div class="data-item">Darkness Condition: ${data.darknessCondition}</div>
                <div class="data-item">Alert Status: ${data.alertStatus}</div>
            ` + "`" + `;
        }

        updateWeatherData();
        setInterval(updateWeatherData, pollingInterval);
    </script>
</body>
</html>
`

	t, err := template.New("home").Parse(tmpl)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		PollingInterval int
	}{
		PollingInterval: int(getPollingIntervalMilliseconds()),
	}

	w.Header().Set("Content-Type", "text/html")
	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		// Note: At this point, it's too late to call http.Error because we've already started writing the response
	}
}

func getPollingIntervalMilliseconds() int64 {
	duration, err := time.ParseDuration(config.PollingInterval)
	if err != nil {
		log.Printf("Error parsing polling interval: %v", err)
		return 60000 // Default to 1 minute if there's an error
	}
	return duration.Milliseconds()
}

func handleDescription(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"Value": "Boltwood II Weather Data Driver",
	})
}

func handleDriverInfo(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"Value": "ASCOM Alpaca Boltwood II Weather Data Driver v0.1",
	})
}

func handleDriverVersion(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{
		"Value": "v0.1",
	})
}

func handleTemperature(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the common query parameters
	clientID := r.URL.Query().Get("ClientID")
	clientTransactionIDStr := r.URL.Query().Get("ClientTransactionID")

	// Convert ClientTransactionID to uint32
	clientTransactionID, err := strconv.ParseUint(clientTransactionIDStr, 10, 32)
	if err != nil {
		clientTransactionID = 0 // Default to 0 if parsing fails
	}

	// Prepare the response structure
	response := struct {
		Value               float64 `json:"Value"`
		ClientTransactionID uint32  `json:"ClientTransactionID"`
		ServerTransactionID uint32  `json:"ServerTransactionID"`
		ErrorNumber         int     `json:"ErrorNumber"`
		ErrorMessage        string  `json:"ErrorMessage"`
	}{
		ClientTransactionID: uint32(clientTransactionID),
		ServerTransactionID: uint32(getNextTransactionID()),
		ErrorNumber:         0,
		ErrorMessage:        "",
	}

	if r.Method != "GET" {
		response.ErrorNumber = 1007 // Invalid Operation
		response.ErrorMessage = "Method not allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	response.Value = weatherData.SensorTemperature

	// Log the request (optional, but helpful for debugging)
	log.Printf("Temperature request: ClientID=%s, ClientTransactionID=%d, Value=%v",
		clientID, clientTransactionID, response.Value)

	json.NewEncoder(w).Encode(response)
}

func handleHumidity(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Value": weatherData.Humidity,
	})
}

func handleWeatherAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weatherData)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	// Implement status endpoint
}

func handleWeather(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(weatherData)
}

func handleAPIVersions(w http.ResponseWriter, r *http.Request) {
	versions := []int{1}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Value": versions,
	})
}

func handleConfiguredDevices(w http.ResponseWriter, r *http.Request) {
	devices := []map[string]interface{}{
		{
			"DeviceName":   "Boltwood II Weather Station",
			"DeviceType":   "ObservingConditions",
			"DeviceNumber": 0,
			"UniqueID":     generateUniqueID(),
		},
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Value": devices,
	})
}

func handleManagementDescription(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()

	description := struct {
		ServerName          string `json:"ServerName"`
		Manufacturer        string `json:"Manufacturer"`
		ManufacturerVersion string `json:"ManufacturerVersion"`
		Location            string `json:"Location"`
		//ServerVersion         string   `json:"ServerVersion"`
		//ServerAPIVersion      string   `json:"ServerAPIVersion"`
		//ServerCodeVersions    []string `json:"ServerCodeVersions"`
		//ServerOperatingSystem string   `json:"ServerOperatingSystem"`
	}{
		ServerName:          "Boltwood II ASCOM Alpaca Server",
		Manufacturer:        "GregOberfield",
		ManufacturerVersion: "0.1",
		Location:            hostname,
		//ServerVersion:         "1.0",
		//ServerAPIVersion:      "1",
		//ServerCodeVersions:    []string{"Alpaca Driver: 0.1", "Boltwood II Adapter: 0.1"},
		//ServerOperatingSystem: runtime.GOOS + " " + runtime.GOARCH,
	}

	// w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Value": description,
	})
	log.Printf("Response: %s", description)
}

// Updated and new device API handlers

func handleConnected(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the common query parameters
	clientID := r.URL.Query().Get("ClientID")
	clientTransactionIDStr := r.URL.Query().Get("ClientTransactionID")

	// Convert ClientTransactionID to uint32
	clientTransactionID, err := strconv.ParseUint(clientTransactionIDStr, 10, 32)
	if err != nil {
		clientTransactionID = 0 // Default to 0 if parsing fails
	}

	// Prepare the response structure
	response := struct {
		Value               interface{} `json:"Value"`
		ClientTransactionID uint32      `json:"ClientTransactionID"`
		ServerTransactionID int         `json:"ServerTransactionID"`
		ErrorNumber         int         `json:"ErrorNumber"`
		ErrorMessage        string      `json:"ErrorMessage"`
	}{
		ClientTransactionID: uint32(clientTransactionID),
		ServerTransactionID: getNextTransactionID(),
	}

	switch r.Method {
	case "GET":
		connMutex.Lock()
		response.Value = isConnected
		connMutex.Unlock()

	case "PUT":
		// Parse form data
		if err := r.ParseForm(); err != nil {
			response.ErrorNumber = 1001 // Invalid Value
			response.ErrorMessage = "Failed to parse form data"
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Get the Connected value from form data
		connectedStr := r.FormValue("Connected")
		connected, err := strconv.ParseBool(connectedStr)
		if err != nil {
			response.ErrorNumber = 1001 // Invalid Value
			response.ErrorMessage = "Invalid Connected value"
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		connMutex.Lock()
		isConnected = connected
		response.Value = nil // PUT requests should return null for Value
		connMutex.Unlock()

	default:
		response.ErrorNumber = 1007 // Invalid Operation
		response.ErrorMessage = "Method not allowed"
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Log the request (optional, but helpful for debugging)
	log.Printf("Connected request: Method=%s, ClientID=%s, ClientTransactionID=%d, Value=%v",
		r.Method, clientID, clientTransactionID, response.Value)

	json.NewEncoder(w).Encode(response)
}

// Implement this function to generate unique server transaction IDs
var serverTransactionID int
var serverTransactionMutex sync.Mutex

func getNextTransactionID() int {
	serverTransactionMutex.Lock()
	defer serverTransactionMutex.Unlock()
	serverTransactionID++
	return serverTransactionID
}

func handleName(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Value": "Boltwood II Weather Station",
	})
}

func handleSupportedActions(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Value": []string{}, // No supported actions for this simple driver
	})
}

func handleInterfaceVersion(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"Value": 1, // The version of the ObservingConditions interface
	})
}

// Initialize the random number generator with a seed
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// generateUniqueID creates a unique identifier for the Alpaca device
func generateUniqueID() string {
	// Generate a random number between 1 and 65535
	randomNum := rng.Intn(65535) + 1

	// Convert the number to a string
	return fmt.Sprintf("%d", randomNum)
}
