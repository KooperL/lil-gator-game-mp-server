package main


import "flag"
import "log"
import "fmt"
import "net/http"

var addr = flag.String("addr", ":8000", "http service address")

func main() {
    fmt.Println("Server starting")

  	hub := newHub()
  	go hub.run()

  	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
