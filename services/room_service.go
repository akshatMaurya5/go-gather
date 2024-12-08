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
	rs := &RoomService{Room: room}
	go rs.handleRoomEvents() // Start listening to events
	return rs
}

func generateRandomId() string {
	return uuid.New().String()
}

func createRoom() *models.Room {
	roomId := generateRandomId()[:5]
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
	log.Printf("Joining room: %s with User ID: %s", rs.Room.ID, client.UserId)

	if _, exists := rs.Room.Clients[client.UserId]; exists {
		log.Printf("User %s is already in room %s", client.UserId, rs.Room.ID)
		return
	}

	rs.Room.Clients[client.UserId] = client
	rs.Room.Join <- []*models.Socket{client}
	log.Printf("User %s successfully joined room %s", client.UserId, rs.Room.ID)
	log.Printf("Current participants in room %s: %d", rs.Room.ID, len(rs.Room.Clients))
}

func (rs *RoomService) handleRoomEvents() {
	for {
		select {
		case clients := <-rs.Room.Join:
			for _, client := range clients {
				log.Printf("User %s joined room %s", client.UserId, rs.Room.ID)
			}

		case message := <-rs.Room.Broadcast:
			for _, client := range rs.Room.Clients {
				client.Conn.WriteMessage(1, message)
			}

		case clients := <-rs.Room.Leave:
			for _, client := range clients {
				delete(rs.Room.Clients, client.UserId)
				log.Printf("User %s left room %s", client.UserId, rs.Room.ID)
			}
		}
	}
}
