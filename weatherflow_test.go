package weatherflow_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/tris/weatherflow"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func startMockServer() (string, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", mockServerHandler)
	server := &http.Server{
		Handler: mux,
	}

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	go func() {
		err := server.Serve(ln)
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	port := ln.Addr().(*net.TCPAddr).Port
	wsURL := fmt.Sprintf("ws://localhost:%d/ws", port)

	stopServer := func() {
		server.Shutdown(context.Background())
	}

	return wsURL, stopServer
}

func mockServerHandler(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		panic(err)
	}
	defer c.Close(websocket.StatusInternalError, "Internal error")

	// Send connection_opened message
	openMsg := map[string]string{"type": "connection_opened"}
	if err := wsjson.Write(r.Context(), c, openMsg); err != nil {
		return
	}

	for {
		var msg map[string]interface{}
		err := wsjson.Read(r.Context(), c, &msg)
		if err != nil {
			return
		}

		switch msg["type"].(string) {
		case "listen_start":
			// Send ack and obs_st messages
			_ = wsjson.Write(r.Context(), c, map[string]string{"type": "ack", "id": msg["id"].(string)})
			_ = wsjson.Write(r.Context(), c, map[string]interface{}{
				"status": map[string]interface{}{
					"status_code":    0,
					"status_message": "SUCCESS",
				},
				"device_id": 121037,
				"type":      "obs_st",
				"source":    "cache",
				"summary": map[string]interface{}{
					"pressure_trend":                     "steady",
					"strike_count_1h":                    0,
					"strike_count_3h":                    0,
					"precip_total_1h":                    0.0,
					"strike_last_dist":                   38,
					"strike_last_epoch":                  1679435903,
					"precip_accum_local_yesterday":       0.0,
					"precip_accum_local_yesterday_final": 0.0,
					"precip_analysis_type_yesterday":     0,
					"raining_minutes":                    []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
					"precip_minutes_local_day":           0,
					"precip_minutes_local_yesterday":     0,
				},
				"obs": [][]interface{}{
					{
						1681767864, 4.19, 4.24, 4.27, 285, 20, 722.7, nil, nil, 109435, 6.19, 912, 0, 0, 0, 0, 2.46, 1, 0, 0, 0, 0,
					},
				},
			})

		case "listen_rapid_start":
			// Send ack and rapid_wind messages
			_ = wsjson.Write(r.Context(), c, map[string]string{"type": "ack", "id": msg["id"].(string)})
			_ = wsjson.Write(r.Context(), c, map[string]interface{}{
				"type":          "rapid_wind",
				"device_id":     121037,
				"serial_number": "ST-00026524",
				"hub_sn":        "HB-00039816",
				"ob": []interface{}{
					1681768025, 4.27, 282,
				},
			})
		}
	}
}

func TestNewClient(t *testing.T) {
	// Start a local WebSocket server for testing
	url, stopServer := startMockServer()
	defer stopServer()

	// Create client
	client := weatherflow.NewClient("your_token", t.Logf)
	client.AddDevice(12345)

	// Override URL to point to our mock server
	client.SetURL(url)

	// Create a channel to receive messages
	msgCh := make(chan weatherflow.Message)

	// Start the client
	client.Start(func(msg weatherflow.Message) {
		msgCh <- msg
	})

	// Use a select statement with a timeout to check if the expected messages are received
	timeout := 5 * time.Second
	expectedMessages := 2

	for i := 0; i < expectedMessages; i++ {
		select {
		case msg := <-msgCh:
			t.Logf("Received message: %#v", msg)
			// TODO: add specific checks for the received messages
		case <-time.After(timeout):
			t.Fatalf("Timed out waiting for message %d", i+1)
		}
	}

	// Stop the client
	client.Stop()
}
