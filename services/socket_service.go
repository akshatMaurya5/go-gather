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
	rooms    map[string]*RoomService
}

func NewSocketService() *SocketService {
	return &SocketService{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin:     func(r *http.Request) bool { return true },
		},
		rooms: make(map[string]*RoomService),
	}
}

func (s *SocketService) Start() {
	http.HandleFunc("/ws", s.handleWebSocket)
	log.Println("Socket service running on port 64527")

	if err := http.ListenAndServe(":64527", nil); err != nil {
		log.Fatalf("Error starting socket service: %v", err)
	}
}

func (s *SocketService) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	clientAddr := wsConn.RemoteAddr().String()
	s.clients.Store(clientAddr, wsConn)
	log.Println("Client connected:", clientAddr)

	go s.handleClient(wsConn, clientAddr)
}

func (s *SocketService) handleClient(wsConn *websocket.Conn, clientAddr string) {
	defer func() {
		s.clients.Delete(clientAddr)
		wsConn.Close()
		log.Println("Client disconnected:", clientAddr)
	}()

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

		log.Printf("Received message from %s: %+v", clientAddr, msg)

		switch msg.Type {
		case "createRoom":
			room := createRoom()
			s.rooms[room.ID] = NewRoomService(room)
			s.BroadcastMessage(Message{Type: "info", Content: fmt.Sprintf("Room created with ID: %s", room.ID)})
		case "joinRoom":
			var joinData struct {
				RoomId string `json:"roomId"`
				UserId string `json:"userId"`
			}
			if err := json.Unmarshal([]byte(msg.Content), &joinData); err == nil {
				roomService, exists := s.rooms[joinData.RoomId]
				if exists {
					client := &models.Socket{Conn: wsConn, UserId: joinData.UserId, RoomId: joinData.RoomId}
					roomService.Join(client)
				} else {
					log.Printf("Room %s does not exist", joinData.RoomId)
				}
			} else {
				log.Printf("Error unmarshalling join data: %v", err)
			}
		case "chat":
			var chatData struct {
				RoomId  string `json:"roomId"`
				UserId  string `json:"userId"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal([]byte(msg.Content), &chatData); err == nil {
				roomService, exists := s.rooms[chatData.RoomId]
				if exists {
					chatMsg := Message{
						Type:    "chat",
						Content: fmt.Sprintf("%s: %s", chatData.UserId, chatData.Message),
					}

					// Broadcast to all room participants
					for _, client := range roomService.Room.Clients {
						err := client.Conn.WriteJSON(chatMsg)
						if err != nil {
							log.Printf("Error sending chat to %s: %v", client.UserId, err)
						}
					}
					log.Printf("Broadcasted message to room %s: %s", chatData.RoomId, chatMsg.Content)
				} else {
					log.Printf("Room %s does not exist for chat", chatData.RoomId)
				}
			} else {
				log.Printf("Error unmarshalling chat data: %v", err)
			}

		default:
			log.Printf("Unknown message type: %s", msg.Type)
		}
	}
}

func (s *SocketService) BroadcastMessage(msg Message) {
	s.clients.Range(func(key, value interface{}) bool {
		client := value.(*websocket.Conn)
		if err := client.WriteJSON(msg); err != nil {
			log.Printf("Error broadcasting to %v: %v", key, err)
			s.clients.Delete(key)
		}
		return true
	})
}

func TestSocketService() {
	socketService := NewSocketService()

	go func() {
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
