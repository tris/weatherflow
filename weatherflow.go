// Package weatherflow provides a client for accessing WeatherFlow's Smart
// Weather API over a WebSocket connection.
package weatherflow

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	wfURL          = "wss://ws.weatherflow.com/swd/data?token=%s"
	initialBackoff = 2 // seconds (don't set below 2)
	maxBackoff     = 32
)

var (
	defaultTimeout = 12 * time.Hour
)

// Client represents a client for the WeatherFlow Smart Weather API.
type Client struct {
	deviceIDs map[int]struct{}
	url       string
	timeout   time.Duration
	logf      Logf
	conn      *websocket.Conn
	errors    int
	ready     bool
	ctx       context.Context
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// NewClient creates a new Client with the given API token, optional connection
// timeout, and an optional log function (if nil, logs will be discarded).
func NewClient(token string, timeout *time.Duration, logf Logf) *Client {
	if logf == nil {
		logf = func(format string, args ...interface{}) {} // discard
	}

	if timeout == nil {
		timeout = &defaultTimeout
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		deviceIDs: make(map[int]struct{}),
		url:       fmt.Sprintf(wfURL, token),
		timeout:   *timeout,
		logf:      logf,
		ctx:       ctx,
		cancel:    cancel,
	}

	return c
}

// SetURL overrides the server URL (for testing).
func (c *Client) SetURL(url string) {
	c.url = url
}

// AddDevice subscribes to wind events for a device ID.
func (c *Client) AddDevice(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deviceIDs[id] = struct{}{}

	if c.conn != nil && c.ready {
		c.sendListenStart(id)
	}
}

// RemoveDevice unsubscribes from wind events for a device ID.
func (c *Client) RemoveDevice(id int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.deviceIDs, id)

	if c.conn != nil && c.ready {
		c.sendListenStop(id)
	}
}

// DeviceCount returns a count of monitored devices.
func (c *Client) DeviceCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.deviceIDs)
}

// Start initiates a WebSocket connection to the WeatherFlow server and processes
// incoming messages.
func (c *Client) Start(onMessage func(Message)) {
	go func() {
		defer c.cancel()

		for {
			select {
			case <-c.ctx.Done():
				// Close the WebSocket connection and return
				// Probably redundant?
				if c.conn != nil {
					c.logf("Disconnecting from WeatherFlow")
					_ = c.conn.Close(websocket.StatusNormalClosure, "Closing connection")
				}
				c.ready = false
				return

			default:
				c.handleBackoff()
				c.logf("Connecting to WeatherFlow")
				conn, _, err := websocket.Dial(c.ctx, c.url, nil)
				if err != nil {
					c.logf("Error connecting to WeatherFlow: %v", err)
					c.errors++
					continue
				}

				// Start a ticker for the connection timeout
				ticker := time.NewTicker(c.timeout)
				defer ticker.Stop()

				defer conn.Close(websocket.StatusInternalError, "closing connection")
				c.conn = conn

				// Read messages from the WebSocket connection
				readLoop:
				for {
					select {
					case <-ticker.C:
						c.logf("Connection timeout")
						break readLoop

					default:
						msgType, msg, err := conn.Read(c.ctx)
						if err != nil {
							if !errors.Is(err, context.Canceled) {
								c.logf("Error reading message: %v", err)
								c.errors++
							}
							break readLoop
						}

						if msgType != websocket.MessageText {
							c.logf("Error resolving unexpected message type: %v", msgType)
							c.errors++
							continue
						}

						// Parse the message
						m, err := UnmarshalMessage(msg)
						if err != nil {
							c.logf("Error unmarshalling message: %v", err)
							c.errors++
							continue
						}

						// Handle the message
						switch t := m.(type) {
						case *MessageRapidWind:
							onMessage(m)

						case *MessageObsSt:
							onMessage(m)

						case *MessageAck:
							c.logf("Received ack: %s", t.ID)

						case *MessageConnectionOpened:
							// Subscribe to wind events
							c.mu.Lock()
							c.ready = true
							for id, _ := range c.deviceIDs {
								c.sendListenStart(id)
							}
							c.mu.Unlock()

						default:
							c.logf("Received unknown message: %v", t)
						}

						// One good message resets the error counter.
						// Set to 1 to enforce minimum backoff between reconnects.
						c.errors = 1
					}
				}
			}
		}
	}()
}

// handleBackoff sleeps for up to maxBackoff seconds to avoid overwhelming
// the API when it's having issues.
func (c *Client) handleBackoff() {
	// No backoff if we haven't gotten any errors yet.
	if c.errors == 0 {
		return
	}

	backoff := math.Min(math.Pow(initialBackoff, float64(c.errors)), maxBackoff)
	c.logf("sleeping for %.0f sec after %d error(s)", backoff, c.errors)
	time.Sleep(time.Duration(backoff) * time.Second)
}

// sendListenStart subscribes to wind observation events.
func (c *Client) sendListenStart(id int) {
	c.logf("Listening to wind events from device %d", id)

	idStr := strconv.Itoa(id)

	startMessage := map[string]interface{}{
		"type":      "listen_start",
		"device_id": id,
		"id":        "listen_start_" + idStr,
	}

	rapidStartMessage := map[string]interface{}{
		"type":      "listen_rapid_start",
		"device_id": id,
		"id":        "listen_rapid_start_" + idStr,
	}

	err := wsjson.Write(c.ctx, c.conn, startMessage)
	if err != nil {
		c.logf("Error sending start message: %v", err)
		c.errors++
	}

	err = wsjson.Write(c.ctx, c.conn, rapidStartMessage)
	if err != nil {
		c.logf("Error sending rapid start message: %v", err)
		c.errors++
	}
}

// sendListenStop unsubscribes from wind observation events.
func (c *Client) sendListenStop(id int) {
	c.logf("Stopping wind events from device %d", id)

	idStr := strconv.Itoa(id)

	stopMessage := map[string]interface{}{
		"type":      "listen_stop",
		"device_id": id,
		"id":        "listen_stop_" + idStr,
	}

	rapidStopMessage := map[string]interface{}{
		"type":      "listen_rapid_stop",
		"device_id": id,
		"id":        "listen_rapid_stop_" + idStr,
	}

	err := wsjson.Write(c.ctx, c.conn, stopMessage)
	if err != nil {
		c.logf("Error sending stop message: %v", err)
		c.errors++
	}

	err = wsjson.Write(c.ctx, c.conn, rapidStopMessage)
	if err != nil {
		c.logf("Error sending rapid stop message: %v", err)
		c.errors++
	}
}

func (c *Client) Stop() {
	c.cancel()
}
