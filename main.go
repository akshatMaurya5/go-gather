package main

import (
	"log"
	"net/http"

	"go-gather/ws"

	"github.com/gorilla/websocket"
)

func main() {
	log.Print("Starting WebSocket server...")

	wsh := ws.WebSocketHandler{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	http.HandleFunc("/ws", wsh.ServeHTTP)

	log.Print("WebSocket server listening on port 3001...")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
