package main

import (
	"fmt"
	"maps"
	"slices"
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
			fmt.Println("Registering new client")
			if _, ok := h.clientConnectionPool[client.sessionKey]; !ok {
				h.clientConnectionPool[client.sessionKey] = map[*Client]PlayerData{client: PlayerData{}}
			} else {
				h.clientConnectionPool[client.sessionKey][client] = PlayerData{}
			}

		case client := <-h.unregister:
			fmt.Println("Unregistering client")
			if _, ok := h.clientConnectionPool[client.sessionKey][client]; ok {
				delete(h.clientConnectionPool[client.sessionKey], client)
				close(client.send)
			}

		case client := <-h.updatePlayerData:
			h.clientConnectionPool[client.sessionKey][client] = client.latestPlayerData

		case _ = <-h.broadcast:
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
					client.send <- payloadAsBytes

				}
			}
			// fmt.Printf("Broadcasted to %d sessions with %d total clients.\n", len(h.clientConnectionPool), totalClientsConnected)
		}
	}
}
