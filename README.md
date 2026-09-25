# Assignment 1: Weather Stations

**Due: Fri, Sep 26th**

Please make sure to regularly commit and push your work to Github.
As with all assignments in this course, 5% of the grade will come from the quality of your git commit history.

# GetWeatherData

http://weatherstation/ is a service that provides temperature readings for weather stations across the world. Since these stations are distributed globally, the service may take some time to respond or may not respond at all.

Your first task is to implement the `GetWeatherData` function in [weather_client.go](weather_client.go), which should make an RPC request to a weather station with a given ID at http://weatherstation:8080/weather?id={id} and return the temperature reading as a float64. 

You can test your solution by running:

```bash
go run weather_client.go
```

Note that in order for your RPC client to work, it will need an RPC server. You can run one by opening up another terminal, navigating to the same directory, and running the server (the first line is the command, the second line is the expected output):

```bash
go run weather_server/weather_server.go
> Weather RPC server running on port 8080
```

The test file will handle the servers for you. You only need to run the server manually when running `weather_client.go` in isolation. To repeat, you *do not* need to run the server when running `go test...`.

# Aggregator

There are `k` weather stations around the world. Your second task is to compute the current mode (the most common) temperature across these stations and also track how many stations reported that value every `averagePeriod` seconds. 

However, as mentioned, the weather station may take some time to respond or may not respond at all. Compute these values for the weather stations that do respond, ensuring the entire calculation is completed within `averagePeriod` seconds, and send the result as a pair [mode, count], where count is the number of stations that reported the mode value to the `out` channel every `averagePeriod` seconds. If two or more temperatures are tied for the mode, go with the lower one. For example, if temperatures 56.4 and 86.5 are both reported by 5 weather stations, pick 56.4 because it is smaller than 86.5. Your implementation should also gracefully handle shutdown requests from the `quit` channel, i.e. it should terminate immediately upon receiving a signal on the `quit` channel.

<u>**Late responses should be ignored instead of included in the next batch.**</u>

If a batch contains zero observations, return [NaN, 0].

You will be implementing **two** distinct solutions to this problem.

1. **Channel-based solution:** Your first implementation should exclusively utilize channels. In this approach, the use of mutexes or locks is not permitted.
2. **Mutex-based solution:** For your second implementation, you should modify your approach to instead rely on mutexes for managing concurrency.

## Testing your code

We've provided a set of unit tests in the [weatherstation_test.go](weatherstation_test.go) file that you can use to test your code. The tests mock the weather station responses. You can run them using the following command in the `a1-weatherstations` directory: 

```bash
go test -v -race
```

## Submission

Upload your Github repository to Gradescope. Make sure it includes your `weather_client.go`, `channel_aggregator.go`, and `mutex_aggregator.go` files. Please do not change any of the function signatures or the package definition, or the autograder may fail to run.