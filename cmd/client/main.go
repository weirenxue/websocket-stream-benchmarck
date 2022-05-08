package main

import (
	"fmt"
	"log"
	"sync"
	"time"
	"websocket-steam-benchmark/util"

	"github.com/c2h5oh/datasize"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

func main() {
	// Load config from config.toml
	config, err := util.LoadConfig("config")
	if err != nil {
		log.Fatal("loadconfig:", err)
	}

	// Two maps for storing the timestamp and duration of the request.
	var beginTimestamp = make(map[string]time.Time)
	var duration = make(map[string]time.Duration)
	var mux sync.RWMutex

	// Testing parameters, requestDuration for each request
	// interval and totalPackage for the total requests that
	// should be sent.
	var requestDuration = config.Client.RequestDuration
	var totalPackage = config.Client.TotalRequest

	// Connect to the server.
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+config.Server.Host+":"+config.Server.Port+"/ws", nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	// Variable that calculates how many
	// dummy bytes are received from server.
	var totalReceivedDummyMessageBytes datasize.ByteSize
	// Read messages from the connection.
	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Fatal("read message:", err)
			}
			mux.Lock()
			// Get timestamp if this is the response to a request.
			t, ok := beginTimestamp[string(message)]
			if !ok {
				// This message is dummy message.
				totalReceivedDummyMessageBytes += datasize.ByteSize(len(message))
			} else {
				// This message is the response to a request.
				// Record the round trip time.
				duration[string(message)] = time.Since(t)
			}
			mux.Unlock()
		}
	}()

	// All messages to the server should be sent to this channel.
	messageCh := make(chan []byte, 1)
	// Write messages to the connection.
	go func() {
		for message := range messageCh {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Fatal("write message:", err)
			}
		}
	}()

	// Send the request message and record the timestamp.
	ticker := time.NewTicker(requestDuration)
	packages := 0
	for range ticker.C {
		packages += 1
		go func() {
			mux.Lock()
			payload := uuid.New().String()
			beginTimestamp[payload] = time.Now()
			mux.Unlock()

			messageCh <- []byte(payload)
		}()
		if packages == totalPackage {
			ticker.Stop()
			break
		}
	}

	// Waiting for the rest of the responses.
	for {
		if len(duration) != totalPackage {
			time.Sleep(200 * time.Millisecond)
		} else {
			break
		}
	}

	// Sum requests round trip time.
	mux.Lock()
	var totalTime time.Duration = 0
	for _, value := range duration {
		totalTime += value
	}
	mux.Unlock()

	// Output metrics.
	fmt.Printf("request packages: %d\n", len(duration))
	fmt.Printf("request time: %v\n", requestDuration)
	fmt.Printf("total duration: %v\n", totalTime)
	fmt.Printf("average duration: %v\n", totalTime/time.Duration(len(duration)))
	fmt.Printf("total received dummy message: %v\n", totalReceivedDummyMessageBytes)
}
