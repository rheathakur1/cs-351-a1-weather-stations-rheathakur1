package main

import (
	"net/rpc"
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
	// TODO: Your code here.
}
