package main

import (
	"net/rpc"
	"math"
	"time"
)

// Channel-based aggregator that reports the global mode temperature periodically.
//
// Report the mode (most common) temperature across all k weather stations every
// averagePeriod seconds, and also include how many stations reported that value,
// by sending [mode, count] to the out channel. The aggregator should terminate
// upon receiving a signal on the quit channel.
//
// Note! To receive credit, channelAggregator must not use mutexes.
func channelAggregator(
	k int,
	averagePeriod float64,
	out chan [2]float64,
	quit chan struct{},
	client *rpc.Client,
) {
	period := time.Duration(averagePeriod * float64(time.Second))

	for {
		responses := make(chan float64, k)
	
		for i := 0; i < k; i++ {
			go func(stationID int) {
				temp, err := GetWeatherData(client, stationID)
					
				if err == nil {
					responses <- temp
				}
			}(i + 1) 
	}

	counts := make(map[float64]int)

	timer := time.NewTimer(period)

	collecting := true

	for collecting {
		select {
		case temp := <-responses:
			counts[temp]++

		case <-timer.C:
			collecting = false

		case <-quit:
			timer.Stop()
			return
		}
	}
	}

	mode := math.NaN()
	maxCount := 0

	for temp, count := range counts {
		if count > maxCount || (count == maxCount && temp < mode) {
			mode = temp
			maxCount = count
		}
	}

	select {
	case out <- [2]float64{mode, float64(maxCount)}:
	case <-quit:
    	return

}
}
