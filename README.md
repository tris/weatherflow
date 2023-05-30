# weatherflow

weatherflow is a Go module for streaming rapid observations from the
[WeatherFlow Tempest API](https://weatherflow.github.io/Tempest/).

## Installation

```
go get -u github.com/tris/weatherflow
```

## Example

```go
import (
	"fmt"
	"log"
	"github.com/tris/weatherflow"
)

func main() {
	client := weatherflow.NewClient("your-token-here", nil, log.Printf)

	client.AddDevice(12345)

	client.Start(func(msg weatherflow.Message) {
		switch m := msg.(type) {
		case *weatherflow.MessageObsSt:
			fmt.Printf("Observation: %+v\n", m)
		case *weatherflow.MessageRapidWind:
			fmt.Printf("Rapid wind: %+v\n", m)
		}
	})

	time.Sleep(30 * time.Second)

	client.Stop()
}
```

## Limitations

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
