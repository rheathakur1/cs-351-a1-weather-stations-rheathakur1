package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"strconv"
	"time"
)

// WeatherService provides weather-related RPC methods
type WeatherService struct{}

// TemperatureRequest represents an RPC request with a station ID
type TemperatureRequest struct {
	StationID string
}

// TemperatureResponse represents an RPC response with the temperature value
type TemperatureResponse struct {
	Temperature float64
}

// GetTemperature simulates fetching the temperature from a weather station
func (w *WeatherService) GetTemperature(req TemperatureRequest, res *TemperatureResponse) error {
	// Validate station ID
	if _, err := strconv.Atoi(req.StationID); err != nil {
		return errors.New("invalid station ID")
	}

	// Simulate a random delay (0 to 1 second)
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

	// Generate a random temperature between -30.0 and 50.0
	value := -30.0 + rand.Float64()*(50.0+30.0)
	value = float64(int(value*10)) / 10 // Round to 1 decimal place
	res.Temperature = value
	return nil
}

func main() {
	rand.Seed(time.Now().UnixNano())

	// Create and register the RPC service
	weatherService := new(WeatherService)
	rpc.Register(weatherService)
	port := 8080

	// Listen for incoming RPC connections
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Println("Error starting RPC server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Weather RPC server running on port", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}
		go rpc.ServeConn(conn) // Handle the connection in a goroutine
	}
}
