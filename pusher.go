package pusher

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"time"
)

// Default settings.
// Currently only the ws scheme is supported.
const (
	pusherURL  = "ws://ws.pusherapp.com/app/"
	clientName = "Go-pusher.go"
	version    = "0.0.1"
	protocol   = "7"

	checkWait    = 120 * time.Second
	pongTimeout  = 30 * time.Second
	writeTimeout = 15 * time.Second
)

// Measure how long it's been since the last client activity,
// in order to check if the connection is still alive.
var latestActivity time.Time

// The skeleton of an event sent by Pusher.
// Data is actually a JSON object as well,
// but Pusher encodes it as a string, as explained here:
// http://pusher.com/docs/pusher_protocol#double-encoding
type Event struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type Client struct {
	// The Pusher application key.
	Key string

	// The websocket connection to the Pusher service.
	ws *websocket.Conn

	// This map holds a channel for each event
	// that has defined a callback.
	// We read from the channel to get the associated data
	// and trigger the callback.
	handlers map[string]chan string

	// Messages on this channel will be written to the websocket.
	send chan []byte
}

// Connect to the Pusher server via a websocket,
// and return the Client struct for that connection.
func Connect(key string) (*Client, error) {
	ws, err := startWebSocket(key)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	client := &Client{
		send:     make(chan []byte, 256),
		handlers: make(map[string]chan string),
		ws:       ws,
		Key:      key,
	}

	go client.monitor()
	go client.reader()
	go client.writer()

	return client, nil
}

// Subscribe to the given channel.
// Currently only works for public channels.
func (c *Client) Subscribe(channel string) {
	message, _ := json.Marshal(map[string]interface{}{
		"event": "pusher:subscribe",
		"data": map[string]string{
			"channel": channel,
		},
	})

	c.send <- message
}

// Unsubscribe from the given channel.
func (c *Client) Unsubscribe(channel string) {
	message, _ := json.Marshal(map[string]interface{}{
		"event": "pusher:unsubscribe",
		"data": map[string]string{
			"channel": channel,
		},
	})

	c.send <- message
}

func (c *Client) On(event string, handler func(string)) {
	eventChannel := make(chan string)

	c.handlers[event] = eventChannel

	go func() {
		for {
			data := <-eventChannel

			handler(data)
		}
	}()
}

// Start the websocket connection to Pusher.
// Currently only supports the ws scheme.
func startWebSocket(key string) (*websocket.Conn, error) {
	params := url.Values{}
	params.Set("client", clientName)
	params.Set("version", version)
	params.Set("protocol", protocol)

	url := pusherURL + key + "?" + params.Encode()

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	ws.SetPongHandler(func(string) error {
		latestActivity = time.Now()
		return nil
	})

	return ws, nil
}

// Close the current websocket connection and start a new one.
func (c *Client) reconnect() error {
	c.ws.Close()

	ws, err := startWebSocket(c.Key)

	if err != nil {
		log.Fatal(err)
		return err
	}

	c.ws = ws

	return nil
}

// Check the connection regularly and if it
// has been inactive for too long then attempt
// to send a ping and reconnect if necessary.
func (c *Client) monitor() {
	for {
		time.Sleep(checkWait)

		if !c.isActive() {
			c.ping()

			time.Sleep(pongTimeout)

			if !c.isActive() {
				c.reconnect()
			}
		}
	}
}

func (c *Client) isActive() bool {
	return time.Since(latestActivity) < checkWait
}

func (c *Client) reader() {
	for {
		_, message, err := c.ws.ReadMessage()

		if err != nil {
			log.Fatal(err)
			break
		}

		latestActivity = time.Now()

		c.handleEvent()
	}

	c.ws.Close()
}

// If there's a handler defined for this event,
// then send the associated data to process it.
func (c *Client) handleEvent(payload) {
	message := Event{}

	json.Unmarshal(payload, &message)

	event := message.Event

	if c.handlers[event] != nil {
		c.handlers[event] <- message.Data
	}
}

// Reads messages from the connection's
// send channel, and writes them to the websocket connection.
func (c *Client) writer() {
	for {
		message := <-c.send

		c.write(websocket.TextMessage, message)
	}
}

// Write a message directly to a websocket with
// the given type and payload.
func (c *Client) write(messageType int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeTimeout))

	return c.ws.WriteMessage(messageType, payload)
}

// Send a ping request to the Pusher server.
// FIXME/TODO: handle write errors.
func (c *Client) ping() {
	c.ws.SetWriteDeadline(time.Now().Add(writeTimeout))

	err := c.ws.WriteMessage(websocket.PingMessage, []byte{})

	if err != nil {
		log.Fatal(err)
	}
}
