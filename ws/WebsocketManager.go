package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Client represents a connected user.
type Client struct {
	ID     string
	roomID string
	Conn   *websocket.Conn
}

type Room struct {
	ID      string
	clients map[string]*Client
}

// WebSocketManager manages WebSocket connections.
type WebSocketManager struct {
	rooms map[string]*Room
	lock  sync.RWMutex
}

var instance *WebSocketManager
var once sync.Once

func GetWebSocketInstance() *WebSocketManager {
	once.Do(func() {
		instance = &WebSocketManager{
			rooms: make(map[string]*Room),
		}
	})
	return instance
}

func (ws *WebSocketManager) AddUser(client *Client, roomID string) {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	log.Println("Adding user", client.ID, "to room", roomID)

	if _, exists := ws.rooms[roomID]; !exists {
		ws.rooms[roomID] = &Room{
			ID:      roomID,
			clients: make(map[string]*Client),
		}
	}

	client.roomID = roomID

	ws.rooms[roomID].clients[client.ID] = client

	log.Println("User", client.ID, "added to room", roomID)
}

func (ws *WebSocketManager) RemoveUser(clientID, roomID string) {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	log.Println("Trying to remove user", clientID, "from room", roomID)
	room, exists := ws.rooms[roomID]

	if !exists {
		log.Println("Room", roomID, "not found")
		return
	}

	delete(room.clients, clientID)
	log.Println("User", clientID, "removed from room", roomID)
}

func (ws *WebSocketManager) BroadcastToRoom(roomID, message string) {

	log.Println("Broadcasting message", message, "to room", roomID)
	ws.lock.RLock()
	defer ws.lock.RUnlock()

	room, exists := ws.rooms[roomID]

	if !exists {
		log.Println("Room:", roomID, "not found")
		return
	}

	for _, client := range room.clients {
		err := client.Conn.WriteMessage(websocket.TextMessage, []byte(message))

		if err != nil {
			log.Println("Error writing message to client", client.ID, ":", err)
		}
	}
}

// WebSocketHandler is a struct that holds the websocket upgrader.
type WebSocketHandler struct {
	Upgrader websocket.Upgrader
}

// Implement the ServeHTTP method for the WebSocketHandler
func (h WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handleWebsocket(w, r)
}