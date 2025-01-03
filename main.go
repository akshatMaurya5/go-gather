package main

import (
	"log"
	"net/http"

	"go-gather/ws"

	"github.com/gorilla/websocket"
)

func main() {
	log.Print("Starting WebSocket server...")

	// Define WebSocketHandler
	wsh := ws.WebSocketHandler{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	// Register the /ws endpoint for WebSocket connections
	http.HandleFunc("/ws", wsh.ServeHTTP)

	// Start the server on port 3001
	log.Print("WebSocket server listening on port 3001...")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
