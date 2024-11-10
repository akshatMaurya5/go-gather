package services

import (
	"go-meta/models"
	"log"
)

type RoomService struct {
	Room *models.Room
}

func newRoomService(room *models.Room) *RoomService {
	return &RoomService{
		Room: room,
	}
}

func (r *RoomService) Join(socket *models.Socket) {
	r.Room.Join <- []*models.Socket{socket}
	log.Printf("User %s joined room %s", socket.UserId, r.Room.ID)
}

func (r *RoomService) Leave(socket *models.Socket) {
	r.Room.Leave <- []*models.Socket{socket}
	log.Printf("User %s left room %s", socket.UserId, r.Room.ID)
}

func (r *RoomService) BroadcastMessage(message []byte) {
	r.Room.Broadcast <- message
}
