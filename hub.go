// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "fmt"

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
  clients map[string]map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan map[string][]byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan map[string][]byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
      fmt.Println("Registering new client")
      if _, ok := h.clients[client.sessionKey]; !ok {
        h.clients[client.sessionKey] = map[*Client]bool{client: true}
      } else {
			  h.clients[client.sessionKey][client] = true
      }
		case client := <-h.unregister:
      fmt.Println("Unregistering client")
			if _, ok := h.clients[client.sessionKey][client]; ok {
				delete(h.clients[client.sessionKey], client)
				close(client.send)
			}
		case message := <-h.broadcast:
      if valid := validMessage(message["message"]); valid {
        fmt.Printf("Broadcasting valid message: %s\n", string(message["message"]))
        if validRecipients, ok := h.clients[string(message["sessionKey"])]; ok {
          for client := range validRecipients {
			    	select {
			    	  case client.send <- message["message"]:
			    	  default:
			    	  	close(client.send)
			    	  	delete(h.clients[client.sessionKey], client)
			    	  }
			    }
        }
      } else {
        fmt.Printf("Dropping message: %s\n", string(message["message"]))
      }
		}
	}
}
