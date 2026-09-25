package main

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/rpc"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func (m *WeatherService) GetTemperature(req TemperatureRequest, res *TemperatureResponse) error {
	return m.HandleTemperatureRequest(req, res)
}

// WeatherService simulates the RPC weather service.
type WeatherService struct {
	HandleTemperatureRequest func(req TemperatureRequest, res *TemperatureResponse) error
}

func defaultTemperatureHandler(req TemperatureRequest, res *TemperatureResponse) error {
	id, err := strconv.Atoi(req.StationID)
	if err != nil {
		return errors.New("invalid station ID")
	}

	time.Sleep(200 * time.Millisecond)

	if id == 1 {
		select {}
	}

	var temp float64
	switch {
	case id <= 3:
		temp = 2.0
	case id <= 6:
		temp = 8.0
	default:
		temp = float64(id) * 2.0
	}
	res.Temperature = temp
	return nil
}

func customTemperatureHandlerNoResponse(req TemperatureRequest, res *TemperatureResponse) error {
	select {}
}

func customTemperatureHandlerDelay(req TemperatureRequest, res *TemperatureResponse) error {
	id, err := strconv.Atoi(req.StationID)
	if err != nil {
		return errors.New("invalid station ID")
	}

	time.Sleep(200 * time.Millisecond)

	if id > 7 {
		time.Sleep(400 * time.Millisecond)
	}

	var temp float64
	switch {
	case id <= 1:
		temp = 2.0
	case id <= 5:
		temp = 8.0
	default:
		temp = float64(id) * 2.0
	}
	res.Temperature = temp
	return nil
}

// TestMain sets up the mock RPC server and runs the tests.
func TestMain(m *testing.M) {
	// Give the server some time to start
	time.Sleep(500 * time.Millisecond)

	// Run the tests
	code := m.Run()

	// Exit with the code from m.Run()
	os.Exit(code)
}

func startMockRPCServer(shutdownChan chan struct{}, port int, handler func(req TemperatureRequest, res *TemperatureResponse) error) {
	// Create a new RPC server instance
	rpcServer := rpc.NewServer()

	mockService := &WeatherService{
		HandleTemperatureRequest: handler,
	}

	rpcServer.Register(mockService)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-shutdownChan:
					return
				default:
					continue
				}
			}
			go rpcServer.ServeConn(conn)
		}
	}()

	// Wait for a shutdown signal
	<-shutdownChan
	listener.Close()
}

// getTestRPCClient connects to the mock RPC server.
func getTestRPCClient(port int) (*rpc.Client, error) {
	time.Sleep(1 * time.Second)
	return rpc.Dial("tcp", fmt.Sprintf("localhost:%d", port))
}

func TestGetWeatherData(t *testing.T) {
	// Start the mock RPC server
	shutdownChan := make(chan struct{})
	go startMockRPCServer(shutdownChan, 51000, defaultTemperatureHandler)
	defer close(shutdownChan)

	client, err := getTestRPCClient(51000)
	if err != nil {
		t.Fatalf("Failed to connect to test RPC server: %v", err)
	}
	defer client.Close()

	value, err := GetWeatherData(client, 2)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if reflect.TypeOf(value).Kind() != reflect.Float64 {
		t.Fatalf("Expected float64, got %v", reflect.TypeOf(value).Kind())
	}
}

func TestChannelAggregator(t *testing.T) {
	runTests(t, channelAggregator)
}

func TestMutexAggregator(t *testing.T) {
	runTests(t, mutexAggregator)
}

// runTests executes the set of tests on a given aggregator function.
func runTests(
	t *testing.T,
	aggregatorFunc func(int, float64, chan [2]float64, chan struct{}, *rpc.Client),
) {
	t.Run("TestBasic", func(t *testing.T) {
		testBasic(t, aggregatorFunc)
	})

	t.Run("TestNaN", func(t *testing.T) {
		testNaN(t, aggregatorFunc)
	})

	t.Run("TestBatch", func(t *testing.T) {
		testBatch(t, aggregatorFunc)
	})

	t.Run("TestQuit", func(t *testing.T) {
		testQuit(t, aggregatorFunc)
	})
}

// testBasic tests the basic functionality of the aggregator.
func testBasic(
	t *testing.T,
	aggregatorFunc func(int, float64, chan [2]float64, chan struct{}, *rpc.Client),
) {
	shutdownChan := make(chan struct{})
	go startMockRPCServer(shutdownChan, 51001, defaultTemperatureHandler)
	defer close(shutdownChan)

	client, err := getTestRPCClient(51001)
	if err != nil {
		t.Fatalf("Failed to connect to test RPC server: %v", err)
	}
	defer client.Close()

	out := make(chan [2]float64)
	quit := make(chan struct{})

	margin := 0.05 // 5% margin
	averagePeriod := 0.5
	expectedMode := 2.0
	expectedFrequency := 3

	start := time.Now()
	go aggregatorFunc(10, averagePeriod, out, quit, client)

	for range 4 {
		select {
		case result := <-out:
			modeTemp := result[0]
			count := int(result[1])

			if math.IsNaN(modeTemp) {
				t.Fatalf("ERROR: Expected a value, got NaN")
			}
			if modeTemp != expectedMode {
				t.Fatalf("ERROR: Expected mode is %v, got %v", expectedMode, modeTemp)
			}
			if count != expectedFrequency {
				t.Fatalf("Expected %d stations, got %d", expectedFrequency, count)
			}
			if elapsed := time.Since(start).Seconds(); elapsed < margin*averagePeriod {
				t.Fatalf("ERROR: Time elapsed is not within %d%% of averagePeriod", int(margin*100))
			}
			start = time.Now()
		case <-time.After(time.Duration((averagePeriod+margin*averagePeriod)*1000) * time.Millisecond):
			t.Fatalf("ERROR: Time elapsed is not within %d%% of averagePeriod", int(margin*100))
		}
	}
}

// testNaN tests that a batch with zero observations correctly returns NaN.
func testNaN(
	t *testing.T,
	aggregatorFunc func(int, float64, chan [2]float64, chan struct{}, *rpc.Client),
) {
	shutdownChan := make(chan struct{})
	go startMockRPCServer(shutdownChan, 51002, customTemperatureHandlerNoResponse)
	defer close(shutdownChan)

	client, err := getTestRPCClient(51002)
	if err != nil {
		t.Fatalf("Failed to connect to test RPC server: %v", err)
	}
	defer client.Close()

	out := make(chan [2]float64)
	quit := make(chan struct{})

	margin := 0.05 // 5% margin
	averagePeriod := 0.5

	start := time.Now()
	go aggregatorFunc(1, averagePeriod, out, quit, client)

	select {
	case result := <-out:
		if !(math.IsNaN(result[0]) && int(result[1]) == 0) {
			t.Fatalf("Expected NaN avg and 0 count, got %v", result)
		}
		if time.Since(start).Seconds() < margin*averagePeriod {
			t.Fatalf("ERROR: Time elapsed is not within %d%% of averagePeriod", int(margin*100))
		}
	case <-time.After(time.Duration((averagePeriod+margin*averagePeriod)*1000) * time.Millisecond):
		t.Fatalf("ERROR: Time elapsed is not within %d%% of averagePeriod", int(margin*100))
	}
}

// testBatch tests that the aggregator correctly handles batches of observations.
func testBatch(
	t *testing.T,
	aggregatorFunc func(int, float64, chan [2]float64, chan struct{}, *rpc.Client),
) {
	shutdownChan := make(chan struct{})
	go startMockRPCServer(shutdownChan, 51003, customTemperatureHandlerDelay)
	defer close(shutdownChan)

	client, err := getTestRPCClient(51003)
	if err != nil {
		t.Fatalf("Failed to connect to test RPC server: %v", err)
	}
	defer client.Close()

	out := make(chan [2]float64)
	quit := make(chan struct{})

	margin := 0.05 // 5% margin
	averagePeriod := 0.5
	expectedMode := 8.0
	expectedFrequency := 4

	start := time.Now()
	go aggregatorFunc(10, averagePeriod, out, quit, client)

	for range 3 {
		select {
		case result := <-out:
			modeTemp := result[0]
    		count := int(result[1])
			if math.IsNaN(modeTemp) {
				t.Fatalf("ERROR: Expected a value, got NaN")
			}
			if modeTemp != expectedMode {
				t.Fatalf("ERROR: Expected mode %v, but got %v", expectedMode, modeTemp)
			}
			if count != expectedFrequency {
				t.Fatalf("Expected %d stations, got %d", expectedFrequency, count)
			}
			if elapsed := time.Since(start).Seconds(); elapsed < margin*averagePeriod {
				t.Fatalf("ERROR: Time elapsed is not within %d%% of averagePeriod", int(margin*100))
			}
		case <-time.After(time.Duration((averagePeriod+margin*averagePeriod)*1000) * time.Millisecond):
			t.Fatalf("ERROR: Time elapsed is not within %d%% of averagePeriod", int(margin*100))
		}
	}
}

// testQuit tests that the aggregator correctly quits upon receiving a signal on the quit channel.
func testQuit(
	t *testing.T,
	aggregatorFunc func(int, float64, chan [2]float64, chan struct{}, *rpc.Client),
) {

	shutdownChan := make(chan struct{})
	go startMockRPCServer(shutdownChan, 51004, defaultTemperatureHandler)
	defer close(shutdownChan)

	client, err := getTestRPCClient(51004)
	if err != nil {
		t.Fatalf("Failed to connect to test RPC server: %v", err)
	}
	defer client.Close()

	out := make(chan [2]float64)
	quit := make(chan struct{})

	averagePeriod := 0.5

	go aggregatorFunc(10, averagePeriod, out, quit, client)

	select {
	case _ = <-out:
		// Received first value, now set up the quit timer.
	case <-time.After(time.Duration((averagePeriod+0.1)*1000) * time.Millisecond):
		t.Fatal("ERROR: Timed out waiting for the first value")
	}

	go func() { quit <- struct{}{} }()

	select {
	case <-out:
		t.Fatal("ERROR: received answer after quit signal was sent")
	case <-time.After(time.Duration(averagePeriod*1000*2) * time.Millisecond):
		// Success
	}
}
