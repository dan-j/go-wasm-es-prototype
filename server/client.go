// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package server

import (
	"github.com/dan-j/go-wasm/core"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	initWait = 10 * time.Second
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan core.Event

	resourceID string

	initialised chan struct{}
}

// ReadPump pumps messages from the websocket connection to the hub.
//
// The application runs ReadPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		var event core.Event
		err := c.conn.ReadJSON(&event)
		// _, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		if err := c.hub.recorder.Record(c.resourceID, event); err != nil {
			log.Println("failed to record event: ", err)
			return
		}
		// message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.hub.broadcast <- event
	}
}

// WritePump pumps messages from the hub to the websocket connection.
//
// A goroutine running WritePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	select {
	case <-c.initialised:
		break
	case <-time.After(initWait):
		log.Println("failed to send init state within timeout")
		return
	}

	for {
		select {
		case event, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(event); err != nil {
				log.Println("failed to write JSON")
			}

			n := len(c.send)
			for i := 0; i < n; i++ {
				if err := c.conn.WriteJSON(event); err != nil {
					log.Println("failed to write JSON")
				}
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) SendInitialState() error {
	defer close(c.initialised)

	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		log.Println("failed to set write deadline on initial state: ", err)
		return err
	}

	state := core.Thing{
		ID: c.resourceID,
	}
	if err := c.hub.recorder.Hydrate(c.resourceID, &state); err != nil {
		log.Println("failed to hydrate state: ", err)
		return err
	}

	e := core.Event{Type: "InitialStateEvent"}
	if err := e.SetData(core.InitialStateEvent{Thing: state}); err != nil {
		return err
	}
	if err := c.conn.WriteJSON(e); err != nil {
		log.Println("failed to write JSON")
		return err
	}

	return nil
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("failed to upgrade connection: ", err)
		return
	}

	prefix, id := path.Split(r.URL.Path)
	if prefix != "/ws/boards/" {
		_ = conn.WriteMessage(websocket.CloseMessage, []byte("Invalid URL, expected /ws/boards/:id"))
		_ = conn.Close()
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan core.Event, 256), resourceID: id, initialised: make(chan struct{})}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.WritePump()
	go client.ReadPump()

	if err := client.SendInitialState(); err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, []byte("Failed to send initial state: " + err.Error()))
		_ = conn.Close()
	}
}
