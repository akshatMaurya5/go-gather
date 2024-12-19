package main

import (
	"go-meta/socketService"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func main() {
	log.Print("Starting server...")
	wsh := socketService.WebSocketHandler{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	http.Handle("/", wsh)
	log.Print("WebSocket handler registered. Listening on port 3001...")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
