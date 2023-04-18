# weatherflow

weatherflow is a Go module for streaming rapid observations from the
[WeatherFlow Tempest API](https://weatherflow.github.io/Tempest/).

## Installation

```
go get github.com/tris/weatherflow
```

## Usage

```go
client, err := weatherflow.NewClient(apiToken, deviceID, log.Printf)
if err != nil {
	panic(err)
}

go client.Start(func(msg weatherflow.Message) {
	switch m := msg.(type) {
	case *weatherflow.MessageObsSt:
		fmt.Printf("Observation: %+v\n", m)
	case *weatherflow.MessageRapidWind:
		fmt.Printf("Rapid wind: %+v\n", m)
	}
})

time.Sleep(30 * time.Second)

client.Stop()
```

## Limitations

- Only one device ID is supported per connection
    - Note that per WeatherFlow's
[WebSocket Reference](https://weatherflow.github.io/Tempest/api/ws.html#other-useful-information),
you should only open one connection

- Only Tempest and Rapid Wind observations are passed:
    - [ ] Acknowledgement (ack)
    - [ ] Rain Start Event (evt_precip)
    - [ ] Lightning Strike Event (evt_strike)
    - [ ] Device Online Event (evt_device_online)
    - [ ] Device Offline Event (evt_device_offline)
    - [ ] Station Online Event (evt_station_online)
    - [ ] Station Offline Event (evt_station_online)
    - [x] Rapid Wind (3 sec) (rapid_wind)
    - [ ] Observation (Air) (obs_air)
    - [ ] Observation (Sky) (obs_sky)
    - [x] Observation (Tempest) (obs_st)

Pull requests are welcome.

## Credit

I took a bit of inspiration from the excellent
[goweatherflow](https://github.com/gregorosaurus/goweatherflow) module, which
you should use instead if your Tempest is on the same LAN.
