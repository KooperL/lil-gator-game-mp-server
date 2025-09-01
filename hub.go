package main

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"slices"
	"time"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients
	clientConnectionPool map[sessionKey]PlayerClientPool

	// Inbound messages from the clients.
	broadcast chan bool

	updatePlayerData chan *Client

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:            make(chan bool),
		updatePlayerData:     make(chan *Client),
		register:             make(chan *Client),
		unregister:           make(chan *Client),
		clientConnectionPool: make(map[sessionKey]PlayerClientPool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			now := time.Now()
			if _, ok := h.clientConnectionPool[client.sessionKey]; !ok {
				h.clientConnectionPool[client.sessionKey] = map[*Client]PlayerData{client: PlayerData{}}
			} else {
				h.clientConnectionPool[client.sessionKey][client] = PlayerData{}
			}
			timeElapsed := time.Since(now)
			fmt.Printf("Registering client in %v\n", timeElapsed)

		case client := <-h.unregister:
			now := time.Now()
			if _, ok := h.clientConnectionPool[client.sessionKey][client]; ok {
				delete(h.clientConnectionPool[client.sessionKey], client)
				close(client.send)
			}
			timeElapsed := time.Since(now)
			fmt.Printf("Unregistering client in %v\n", timeElapsed)

		case client := <-h.updatePlayerData:
			now := time.Now()
			messageAsStruct, err := PlayerData{}.fromJSONBytes(client.latestPlayerData)
			if err != nil {
				return
			}
			h.clientConnectionPool[client.sessionKey][client] = messageAsStruct
			client.lastValidUpdate = time.Now()
			if rand.IntN(100) < 2 {
				timeElapsed := time.Since(now)
				fmt.Printf("RANDOM - Updating player data in %v\n", timeElapsed)
			}

		case _ = <-h.broadcast:
			now := time.Now()
			totalClientsConnected := 0

			for sessionKey, playerClientPool := range h.clientConnectionPool {
				totalClientsConnected += len(playerClientPool)

				PlayerDataAsSlice := slices.Collect(maps.Values(playerClientPool))

				payload := PlayerDataHub{
					PlayerStates:  PlayerDataAsSlice,
					ServerVersion: LGGMP_Ver,
				}
				payloadAsBytes, err := payload.toJSONBytes()
				if err != nil {
					fmt.Printf("Error serializing '%s' payload: %v\n", sessionKey, err)
					continue
				}
				for client := range playerClientPool {
					if client.lastValidUpdate.Before(time.Now().Add(-1 * time.Minute)) {
						client.hub.unregister <- client
						continue
					}
					if client.connectedAt.Before(time.Now().Add(-30 * time.Minute)) {
						client.hub.unregister <- client
						continue
					}
					client.send <- payloadAsBytes
				}
			}
			if rand.IntN(100) < 2 {
				timeElapsed := time.Since(now)
				fmt.Printf("RANDOM - Broadcast to %d players in %d sessions took %v\n", totalClientsConnected, len(h.clientConnectionPool), timeElapsed)
			}
		}
	}
}
