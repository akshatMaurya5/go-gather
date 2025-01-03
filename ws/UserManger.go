package ws

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Extract userId and roomId from query parameters
	userID := r.URL.Query().Get("userId")
	roomID := r.URL.Query().Get("roomId")

	if userID == "" || roomID == "" {
		log.Println("userId or roomId is missing")
		conn.WriteMessage(websocket.TextMessage, []byte("userId and roomId are required"))
		return
	}

	// Get WebSocket manager singleton
	wsManager := GetWebSocketInstance()

	// Create client and add to room
	client := &Client{
		ID:   userID,
		Conn: conn,
	}
	wsManager.AddUser(client, roomID)

	// Handle messages from client
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from client %s: %v\n", userID, err)
			wsManager.RemoveUser(userID, roomID)
			break
		}

		// Broadcast message to room
		fmt.Printf("Message from %s in room %s: %s\n", userID, roomID, message)
		wsManager.BroadcastToRoom(roomID, fmt.Sprintf("%s: %s", userID, message))
	}
}
