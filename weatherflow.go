// Package weatherflow provides a client for accessing WeatherFlow's Smart
// Weather API over a WebSocket connection.
package weatherflow

import (
	"context"
	"fmt"
	"math"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	wfURL          = "wss://ws.weatherflow.com/swd/data?token=%s"
	initialBackoff = 1	// seconds
	maxBackoff     = 32
)

// Client represents a client for the WeatherFlow Smart Weather API.
type Client struct {
	deviceIDs []int
	url      string
	logf     Logf
	stopCh   chan struct{}
	conn     *websocket.Conn
	errors   int
}

// NewClient creates a new Client with the given API token, device IDs, and an
// optional log function (if nil, logs will be discarded).
func NewClient(token string, deviceIDs []int, logf Logf) (*Client, error) {
	if logf == nil {
		logf = func(format string, args ...interface{}) {} // discard
	}
	client := &Client{
		deviceIDs: deviceIDs,
		url:      fmt.Sprintf(wfURL, token),
		logf:     logf,
		stopCh:   make(chan struct{}),
	}

	return client, nil
}

// SetURL overrides the server URL (for testing).
func (c *Client) SetURL(url string) {
	c.url = url
}

// handleBackoff sleeps for up to maxBackoff seconds to avoid overwhelming
// the API when it's having issues.
func (c *Client) handleBackoff() {
	// No backoff if we haven't gotten any errors yet.
	if c.errors == 0 {
		return
	}

	backoff := math.Min(math.Pow(initialBackoff, float64(c.errors)), maxBackoff)
	c.logf("sleeping for %d sec after %d error(s)", backoff, c.errors)
	time.Sleep(time.Duration(backoff) * time.Second)
}

// sendListenStart subscribes to wind observation events.
func (c *Client) sendListenStart(ctx context.Context, id int) error {
	startMessage := map[string]interface{}{
		"type":      "listen_start",
		"device_id": id,
		"id":        "",
	}

	rapidStartMessage := map[string]interface{}{
		"type":      "listen_rapid_start",
		"device_id": id,
		"id":        "",
	}

	err := wsjson.Write(ctx, c.conn, startMessage)
	if err != nil {
		return err
	}

	err = wsjson.Write(ctx, c.conn, rapidStartMessage)
	if err != nil {
		return err
	}

	return nil
}

// Start initiates a WebSocket connection to the WeatherFlow server and processes
// incoming messages.
func (c *Client) Start(onMessage func(Message)) {
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			<-c.stopCh
			cancel()
		}()

		for {
			select {
			case <-ctx.Done():
				// Close the WebSocket connection and return
				if c.conn != nil {
					_ = c.conn.Close(websocket.StatusNormalClosure, "Closing connection")
				}
				return

			default:
				c.handleBackoff()
				c.logf("Connecting to WeatherFlow")
				conn, _, err := websocket.Dial(ctx, c.url, nil)
				if err != nil {
					c.logf("Error connecting to WeatherFlow: %v", err)
					c.errors++
					continue
				}
				defer conn.Close(websocket.StatusInternalError, "closing connection")
				c.conn = conn

				// Subscribe to wind events
				for _, id := range c.deviceIDs {
					err = c.sendListenStart(ctx, id)
					if err != nil {
						c.logf("Error sending start message: %v", err)
						c.errors++
						continue
					}
				}

				// Read messages from the WebSocket connection
				for {
					msgType, msg, err := conn.Read(ctx)
					if err != nil {
						c.logf("Error reading message: %v", err)
						c.errors++
						break
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
					if m != nil {
						onMessage(m)
						// One good message resets the error counter.
						// Set to 1 to enforce minimum backoff between reconnects.
						c.errors = 1
					}
				}

				c.logf("Connection lost, attempting to reconnect...")
			}
		}
	}()
}

func (c *Client) Stop() {
	close(c.stopCh)
}
