# Pusher Go Client

![img](http://f.cl.ly/items/2j1i1O2C3o0Q2j2g3w2g/logo.png)

This is an initial draft of a client library for Pusher, written in Go.
So far a few basic features have been implemented, but this is still in no way ready
for production environments.

Check out this README or head over to [Go Walker](https://gowalker.org/github.com/ehq/pusher-go).

## Installation

Install with the `go get` command:

`go get github.com/ehq/pusher-go`

## Example Usage

```go
package main

import (
	"encoding/json"
	"github.com/ehq/pusher-go"
	"log"
)

type Move struct {
	X string `json:"x"`
	Y string `json:"y"`
}

func main() {
	client, err := pusher.Connect("your_pusher_key")

	if err != nil {
		log.Fatal(err)
		return
	}

	client.Subscribe("some_channel")

	moves := make(chan Move)

	client.On("move", func(data string) {
		move := Move{}

		json.Unmarshal([]byte(data), &move)

		moves <- move
	})

	for {
		move := <-moves

		log.Printf("New move has been played at x: %s, y: %s", move.X, move.Y)
	}
}
```

## Next steps

The next goals for the project are:

* Handle disconnects and network problems gracefully
* Support standard [error codes](http://pusher.com/docs/pusher_protocol#error-codes)
  * These also determine if a reconnect should be attempted or not.
* Support [system events](http://pusher.com/docs/pusher_protocol#system-events)
* Encapsulate and refactor the concept of a channel
  * Also include presence channels and private channels (which require authentication)
* Scope event handlers to channels, rather than making them be global.
* Implement [client events](http://pusher.com/docs/pusher_protocol#channel-client-events)
