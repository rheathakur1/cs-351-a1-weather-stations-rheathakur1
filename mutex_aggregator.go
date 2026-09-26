package main

import (
	"net/rpc"
	"math"
	"sync"
	"time"
)

// Mutex-based aggregator that reports the global mode temperature periodically.
//
// Report the mode (most common) temperature across all k weather stations every
// averagePeriod seconds, and also include how many stations reported that value,
// by sending [mode, count] to the out channel. The aggregator should terminate
// upon receiving a signal on the quit channel.
//
// Note! To receive credit, mutexAggregator must implement a mutex-based solution.
func mutexAggregator(
	k int,
	averagePeriod float64,
	out chan [2]float64,
	quit chan struct{},
	client *rpc.Client,
) {
	period := time.Duration(averagePeriod * float64(time.Second))

	for {
		counts := make(map[float64]int)
		var mu sync.Mutex
		
		timer := time.NewTimer(period)

		for i := 0; i < k; i++ {
			go func(stationID int) {
				temp, err := GetWeatherData(client, stationID)
				if err == nil {
					mu.Lock()
					counts[temp]++
					mu.Unlock()
				}
			}(i) // Station IDs are 1-indexed
		}

		select {
		case <-timer.C:
			// Calculate the mode and its count
			mu.Lock()
			mode := math.NaN()
			maxCount := 0
			for temp, count := range counts {
				if count > maxCount || (count == maxCount && temp < mode){
					mode = temp
					maxCount = count
				}
			}
			mu.Unlock()

			select {
			case out <- [2]float64{mode, float64(maxCount)}:
			case <-quit:
				return
			}

		case <-quit:
			timer.Stop()
			return
		}
	}
}
