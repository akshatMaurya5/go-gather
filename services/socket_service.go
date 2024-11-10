package services

import (
	"go-meta/models"
	"log"

	"github.com/gorilla/websocket"
)

type SocketService struct {
	Socket *models.Socket
}

func newSocketService(socket *models.Socket) *SocketService { // creates a new SocketService instance
	return &SocketService{
		Socket: socket,
	}
}
func (s *SocketService) WritePump() {
	defer s.Socket.Conn.Close()

	for msg := range s.Socket.Send {

		if err := s.Socket.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("Write error: %v", err)
			break
		}

	}
}

func (s *SocketService) Close() {
	close(s.Socket.Send)
	s.Socket.Conn.Close()
}
