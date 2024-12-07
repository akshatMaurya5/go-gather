package services

import (
	"fmt"
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
}

func NewSocketService() *SocketService {
	return &SocketService{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for now
			},
		},
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

	// Read messages
	for {
		var msg Message
		err := wsConn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error for %s: %v", clientAddr, err)
			} else {
				log.Printf("Connection closed for %s: %v", clientAddr, err)
			}
			break
		}

		// Validate and log message
		if msg.Type == "" || msg.Content == "" {
			log.Println("Received an empty message from", clientAddr)
		} else {
			log.Printf("Received message from %s: %+v", clientAddr, msg)
		}
	}
}

// Utility function to broadcast message to all connected clients
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
	// Create a new socket service
	socketService := NewSocketService()

	// Simulate some test scenarios
	go func() {
		// Broadcast a test message after a short delay
		time.Sleep(5 * time.Second)
		socketService.BroadcastMessage(Message{
			Type:    "test",
			Content: "Test broadcast message",
		})

		// Simulate periodic test messages
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			socketService.BroadcastMessage(Message{
				Type:    "heartbeat",
				Content: fmt.Sprintf("Server heartbeat at %v", time.Now()),
			})
		}
	}()

	// Additional test logging
	log.Println("Starting WebSocket Test Service")

	// Start the socket service
	socketService.Start()
}
