# Pusher Go Client

![img](http://f.cl.ly/items/2j1i1O2C3o0Q2j2g3w2g/logo.png)

This is an initial draft of a client library for Pusher, written in Go.
So far a few basic features have been implemented, but this is still in no way ready
for production environments.

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
