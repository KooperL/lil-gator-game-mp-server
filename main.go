package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var LGGMP_LISTEN string = ":8000"
var LGGMP_Ver string = "1.0.0"
var LGGMP_TICK_RATE int = 60

var addr = flag.String("addr", LGGMP_LISTEN, "http service address")

func schedule(what func(), delay time.Duration) chan bool {
	stop := make(chan bool)

	go func() {
		for {
			what()
			select {
			case <-time.After(delay):
			case <-stop:
				return
			}
		}
	}()

	return stop
}

func triggerBroadcast(hub *Hub) {
	hub.broadcast <- true
}

func main() {
	fmt.Println(fmt.Sprintf("Server (%s) starting on %s", LGGMP_Ver, LGGMP_LISTEN))

	hub := newHub()
	go hub.run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	_ = schedule(func() {
		triggerBroadcast(hub)
	}, time.Duration(1000/LGGMP_TICK_RATE)*time.Millisecond)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
