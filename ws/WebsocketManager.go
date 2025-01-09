package ws

import (
	"encoding/json"
	"fmt"
	"go-gather/types"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID     string
	roomID string
	Conn   *websocket.Conn
	X      int
	Y      int
}

type Room struct {
	ID      string
	clients map[string]*Client
}

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

type WebSocketHandler struct {
	Upgrader websocket.Upgrader
}

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func createUniqueRoomId() string {
	const charSet = "ABCDEFGHIJKLMNOPQRSTUVWXY0123456789abcdefghijklmnopqrstuvw"

	b := make([]byte, 5)
	for i := range b {
		b[i] = charSet[seededRand.Intn(len(charSet))]
	}
	return string(b)
}

func (ws *WebSocketManager) AddUser(client *Client, roomID string) {
	ws.lock.Lock()
	defer ws.lock.Unlock()

	log.Println("Adding user - wsManager", client.ID, "to room", roomID)

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

// BroadcastMove HAS TO BE REVISITED FOR LOGIC CHECKING
func (ws *WebSocketManager) BroadcastMove(client *Client, roomID string) {
	ws.lock.RLock()
	defer ws.lock.RUnlock()

	_, exists := ws.rooms[roomID]

	if !exists {
		log.Println("Room:", roomID, "not found")
		return
	}

	message := fmt.Sprintf("%s moved to (%d, %d)", client.ID, client.X, client.Y)
	ws.BroadcastToRoom(roomID, message)
}

func (c *Client) SendMessage(eventType string, payload interface{}) {
	response := types.Response{
		Type:    eventType,
		Success: true,
		Data:    payload,
	}

	messageBytes, err := json.Marshal(response)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return
	}

	err = c.Conn.WriteMessage(websocket.TextMessage, messageBytes)
	if err != nil {
		log.Println("Error writing message to client", c.ID, ":", err)
	}
}

func (ws *WebSocketManager) GetUserConnection(userID string) *Client {
	ws.lock.RLock()
	defer ws.lock.RUnlock()

	for _, room := range ws.rooms {
		if client, exists := room.clients[userID]; exists {
			return client
		}
	}
	return nil
}

func (ws *WebSocketManager) GetUsersInRoom(roomID string) []string {
	ws.lock.RLock()
	defer ws.lock.RUnlock()

	room, exists := ws.rooms[roomID]
	if !exists {
		return []string{}
	}

	users := make([]string, 0, len(room.clients))
	for userID := range room.clients {
		users = append(users, userID)
	}
	return users
}

func (h WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handleWebsocket(w, r)
}
