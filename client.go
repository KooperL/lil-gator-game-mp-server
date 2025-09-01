package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type BroadcastKey = string

const (
	UpdateSessionKey  BroadcastKey = "sessionKey"
	UpdateMessage     BroadcastKey = "message"
	UpdateDisplayName BroadcastKey = "displayName"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 8192
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// LGGMP session identifier
	sessionKey sessionKey

	// LGGMP display name
	displayName displayName

	connectedAt time.Time

	lastValidUpdate time.Time

	// LGGMP modded .dll version
	clientVersion string

	// Latest player data message received from this client
	// Mapping playerData to the client in PlayerClientPool is pretty unnecessary now because of this, but I don't care right now
	latestPlayerData []byte

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break

		}
		messageAsBytes := bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.latestPlayerData = messageAsBytes
		c.hub.updatePlayerData <- c
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	queryParams := r.URL.Query()
	sessionKey, sessionKeyOk := queryParams["sessionKey"]
	if !sessionKeyOk {
		fmt.Println("No session key provided")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	clientVersion, clientVersionOk := queryParams["clientVersion"]
	if !clientVersionOk {
		fmt.Println("No client version provided")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	displayName, displayNameOk := queryParams["displayName"]
	if !displayNameOk {
		fmt.Println("No display name provided")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if len(sessionKey) > 1 || len(clientVersion) > 1 || len(displayName) > 1 {
		fmt.Println("Invalid session key or client version")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	if clientVersion[0] != LGGMP_Ver {
		fmt.Println(fmt.Sprintf("Latest version required. Current version: %s", clientVersion[0]))
		http.Error(w, fmt.Sprintf("Latest version required. Current version: %s", clientVersion[0]), http.StatusBadRequest)
		return
	}
	if displayName[0] == "" {
		fmt.Println("Invalid display name")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	for _, session := range hub.clientConnectionPool {
		for _, client := range session {
			if client.DisplayName == displayName[0] {
				fmt.Println("Display name already taken")
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}
	}

	totalClientsConnected := 0
	for session := range hub.clientConnectionPool {
		totalClientsConnected += len(session)
	}

	if totalClientsConnected >= 20 {
		fmt.Println("Server is full")
		http.Error(w, "Server is full", http.StatusBadRequest)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256), sessionKey: sessionKey[0], clientVersion: clientVersion[0], displayName: displayName[0], connectedAt: time.Now(), lastValidUpdate: time.Now()}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
