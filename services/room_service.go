package services

import (
	"go-meta/models"
	"log"

	"github.com/google/uuid"
)

type RoomService struct {
	Room *models.Room
}

func NewRoomService(room *models.Room) *RoomService {
	return &RoomService{
		Room: room,
	}
}

func generateRandomId() string {
	return uuid.New().String()
}

func createRoom() *models.Room {
	roomId := generateRandomId()
	if len(roomId) > 5 {
		roomId = roomId[:5]
	}
	log.Printf("Room created with ID: %s", roomId)

	return &models.Room{
		ID:        roomId,
		Clients:   make(map[string]*models.Socket),
		Broadcast: make(chan []byte),
		Join:      make(chan []*models.Socket),
		Leave:     make(chan []*models.Socket),
	}
}

func (rs *RoomService) Join(client *models.Socket) {
	rs.Room.Clients[client.UserId] = client
	rs.Room.Join <- []*models.Socket{client}
	log.Printf("User %s joined room %s", client.UserId, rs.Room.ID)
	log.Printf("Current participants in room %s: %d", rs.Room.ID, len(rs.Room.Clients))

	// Notify all clients in the room about the new participant
	message := Message{Type: "info", Content: client.UserId + " joined room " + rs.Room.ID}
	rs.Room.Broadcast <- []byte(message.Content) // Send message to broadcast channel
}
