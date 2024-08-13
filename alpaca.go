package main

import (
	"encoding/json"
	"log"
	"net"
	"strings"
)

type AlpacaDiscoveryResponse struct {
	AlpacaPort int    `json:"alpacaPort"`
	Version    int    `json:"version"`
	ID         string `json:"id"`
}

func handleAlpacaDiscovery() {
	addr := net.UDPAddr{
		Port: config.DiscoveryPort,
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		log.Printf("Error listening for UDP packets: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Listening for Alpaca discovery requests on port %d", config.DiscoveryPort)

	for {
		handleNextDiscoveryRequest(conn)
	}
}

func handleNextDiscoveryRequest(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(buffer)
	if err != nil {
		log.Printf("Error reading UDP packet: %v", err)
		return
	}

	receivedData := strings.TrimSpace(string(buffer[:n]))
	log.Printf("Received data from %v: %s", remoteAddr, receivedData)

	isValid := false

	// Check if it's a plain "alpacadiscovery1" string
	if strings.ToLower(receivedData) == "alpacadiscovery1" {
		isValid = true
	} else {
		// Try to parse as JSON
		var discoveryRequest map[string]interface{}
		err = json.Unmarshal(buffer[:n], &discoveryRequest)
		if err == nil {
			// Check if the received message is a valid discovery request
			for key, value := range discoveryRequest {
				if strings.ToLower(key) == "alpacadiscovery1" {
					if intValue, ok := value.(float64); ok && intValue == 1 {
						isValid = true
						break
					}
				}
			}
		}
	}

	if !isValid {
		log.Printf("Received invalid discovery request from %v", remoteAddr)
		return
	}

	log.Printf("Received valid Alpaca discovery request from %v", remoteAddr)

	response := AlpacaDiscoveryResponse{
		AlpacaPort: config.WebServerPort,
		Version:    1,
		ID:         generateUniqueID(),
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON response: %v", err)
		return
	}

	_, err = conn.WriteToUDP(jsonResponse, remoteAddr)
	if err != nil {
		log.Printf("Error sending discovery response: %v", err)
		return
	}

	log.Printf("Sent Alpaca discovery response %s to %v", jsonResponse, remoteAddr)
}
