package services

import (
	"encoding/json"
	"fmt"
	"go-meta/models"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type SocketService struct {
	upgrader websocket.Upgrader
	clients  sync.Map
	rooms    map[string]*RoomService // Track rooms
}

func NewSocketService() *SocketService {
	return &SocketService{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		rooms: make(map[string]*RoomService), // Initialize rooms map
	}
}

func (s *SocketService) Start() {
	http.HandleFunc("/ws", s.handleWebSocket)
	log.Println("Socket service running on port 64527")

	err := http.ListenAndServe(":64527", nil)
	if err != nil {
		log.Fatalf("Error starting socket service: %v", err)
	}
}

func (s *SocketService) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	// Store the connection
	clientAddr := wsConn.RemoteAddr().String()
	s.clients.Store(clientAddr, wsConn)
	log.Println("Client connected:", clientAddr)

	// Handle the connection
	go s.handleClient(wsConn, clientAddr)
}

func (s *SocketService) handleClient(wsConn *websocket.Conn, clientAddr string) {
	defer func() {
		// Cleanup
		s.clients.Delete(clientAddr)
		wsConn.Close()
		log.Println("Client disconnected:", clientAddr)
	}()

	// Send welcome message
	err := wsConn.WriteJSON(Message{Type: "info", Content: "Welcome to the WebSocket server!"})
	if err != nil {
		log.Printf("Error sending welcome message: %v", err)
		return
	}

	for {
		var msg Message
		err := wsConn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Connection closed for %s: %v", clientAddr, err)
			break
		}

		// Log the received message
		log.Printf("Received message from %s: %+v", clientAddr, msg)

		// Handle create room and join room messages
		switch msg.Type {
		case "createRoom":
			room := createRoom()                    // Create a new room
			s.rooms[room.ID] = NewRoomService(room) // Store the room
			s.BroadcastMessage(Message{Type: "info", Content: "Room created with ID: " + room.ID})
			log.Printf("Room created with ID: %s", room.ID) // Log room creation
		case "joinRoom":
			var joinData struct {
				RoomId string `json:"roomId"`
				UserId string `json:"userId"`
			}
			if err := json.Unmarshal([]byte(msg.Content), &joinData); err == nil {
				log.Printf("Join request for Room ID: %s by User ID: %s", joinData.RoomId, joinData.UserId) // Log join request
				roomService, exists := s.rooms[joinData.RoomId]
				if exists {
					client := &models.Socket{Conn: wsConn, UserId: joinData.UserId, RoomId: joinData.RoomId}
					roomService.Join(client)                                               // Join the room
					log.Printf("User %s joined room %s", joinData.UserId, joinData.RoomId) // Log successful join
				} else {
					log.Printf("Room %s does not exist", joinData.RoomId) // Log room not found
				}
			} else {
				log.Printf("Error unmarshalling join data: %v", err) // Log unmarshalling error
			}
		default:
			log.Printf("Unknown message type: %s", msg.Type) // Log unknown message type
		}
	}
}

func (s *SocketService) BroadcastMessage(msg Message) {
	s.clients.Range(func(key, value interface{}) bool {
		client := value.(*websocket.Conn)
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("Error broadcasting to %v: %v", key, err)
			s.clients.Delete(key)
		}
		return true
	})
}

func TestSocketService() {

	socketService := NewSocketService()

	go func() {
		// Broadcast a test message after a short delay
		time.Sleep(5 * time.Second)
		socketService.BroadcastMessage(Message{
			Type:    "test",
			Content: "Test broadcast message",
		})

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			socketService.BroadcastMessage(Message{
				Type:    "heartbeat",
				Content: fmt.Sprintf("Server heartbeat at %v", time.Now()),
			})
		}
	}()

	log.Println("Starting WebSocket Test Service")

	socketService.Start()
}
