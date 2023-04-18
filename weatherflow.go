// Package weatherflow provides a client for accessing WeatherFlow's Smart
// Weather API over a WebSocket connection.
package weatherflow

import (
	"context"
	"fmt"
	"time"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

const (
	wfURL          = "wss://ws.weatherflow.com/swd/data?token=%s"
	initialBackoff = 1 * time.Second
	maxBackoff     = 32 * time.Second
)

// Client represents a client for the WeatherFlow Smart Weather API.
type Client struct {
	token    string
	deviceID int
	url      string
	logf     Logf
	stopCh   chan struct{}
	conn     *websocket.Conn
}

// NewClient creates a new Client with the given API token, device ID, and an
// optional log function (if nil, logs will be discarded).
func NewClient(token string, deviceID int, logf Logf) (*Client, error) {
	if logf == nil {
		logf = func(format string, args ...interface{}) {} // discard
	}
	client := &Client{
		token:    token,
		deviceID: deviceID,
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

// Start initiates a WebSocket connection to the WeatherFlow server and processes
// incoming messages.  Blocks indefinitely, reconnecting as needed.
func (c *Client) Start(onMessage func(Message)) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-c.stopCh
		cancel()
	}()

	backoff := initialBackoff

	for {
		select {
		case <-ctx.Done():
			// Close the WebSocket connection and return
			if c.conn != nil {
				_ = c.conn.Close(websocket.StatusNormalClosure, "Closing connection")
			}
			return ctx.Err()

		default:
			c.logf("connecting to %s", c.url)
			conn, _, err := websocket.Dial(ctx, c.url, nil)
			if err != nil {
				c.logf("Error connecting to WeatherFlow: %v", err)
				backoff = backoff * 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				time.Sleep(backoff)
				continue
			}
			defer conn.Close(websocket.StatusInternalError, "closing connection")
			c.conn = conn

			backoff = initialBackoff

			startMessage := map[string]interface{}{
				"type":      "listen_start",
				"device_id": c.deviceID,
				"id":        "",
			}

			rapidStartMessage := map[string]interface{}{
				"type":      "listen_rapid_start",
				"device_id": c.deviceID,
				"id":        "",
			}

			err = wsjson.Write(ctx, conn, startMessage)
			if err != nil {
				return fmt.Errorf("failed to send start message on WebSocket connection: %v", err)
			}

			err = wsjson.Write(ctx, conn, rapidStartMessage)
			if err != nil {
				return fmt.Errorf("failed to send start message on WebSocket connection: %v", err)
			}

			// Read messages from the WebSocket connection
			for {
				msgType, msg, err := conn.Read(ctx)
				if err != nil {
					return fmt.Errorf("failed to read message from WebSocket connection: %v", err)
				}

				if msgType != websocket.MessageText {
					return fmt.Errorf("received unexpected message type: %v", msgType)
				}

				// Parse the message
				m, err := UnmarshalMessage(msg)
				if err != nil {
					return fmt.Errorf("failed to unmarshal message: %v", err)
				}

				// Handle the message
				if m != nil {
					onMessage(m)
				}
			}

			// not reached?
			c.logf("Connection lost, attempting to reconnect...")
		}
	}
}

func (c *Client) Stop() {
	close(c.stopCh)
}
