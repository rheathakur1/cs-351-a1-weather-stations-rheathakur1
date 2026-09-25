package main

import (
	"log"
	"math"
	"net/rpc"
)

// TemperatureRequest represents an RPC request with a station ID
type TemperatureRequest struct {
	StationID string
}

// TemperatureResponse represents an RPC response with the temperature value
type TemperatureResponse struct {
	Temperature float64
}

// GetWeatherData fetches the temperature reading for a given weather station ID over RPC
func GetWeatherData(client *rpc.Client, id int) (float64, error) {
	// TODO: Your code here

	return math.NaN(), nil
}

// Test GetWeatherData implementation
func main() {
	// Connect to the RPC server
	client, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Error connecting to RPC server:", err)
	}
	defer client.Close()

	// Fetch weather data for station ID 1
	temp, err := GetWeatherData(client, 1)
	if err != nil {
		log.Fatal("Error fetching weather data:", err)
	}

	log.Println("Temperature:", temp)
}
