package socketservice

import (
	"go-meta/models"
	"sync"
)

type RoomManager struct {
	rooms map[string][]*UserSocket
	mutex sync.Mutex
}

var (
	instance *RoomManager
	once     sync.Once
)

func GetInstance() *RoomManager {
	once.Do(func() {
		instance = &RoomManager{
			rooms: make(map[string][]*UserSocket),
		}
	})
	return instance
}

func (rm *RoomManager) AddUser(roomId string, user *UserSocket) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.rooms[roomId]; !exists {
		rm.rooms[roomId] = []*UserSocket{user}
		return
	}
	rm.rooms[roomId] = append(rm.rooms[roomId], user)
}

func (rm *RoomManager) RemoveUser(user *UserSocket, roomId string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.rooms[roomId]; !exists {
		return
	}

	users := rm.rooms[roomId]
	for i, u := range users {
		if u.UserId == user.UserId {
			rm.rooms[roomId] = append(users[:i], users[i+1:]...)
			break
		}
	}
}

func (rm *RoomManager) Broadcast(message *models.Message, sender *UserSocket, roomId string) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if users, exists := rm.rooms[roomId]; exists {
		for _, user := range users {
			if user.UserId != sender.UserId {
				user.Send(message)
			}
		}
	}
}

func (rm *RoomManager) GetUsers(roomId string) []*UserSocket {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	return rm.rooms[roomId]
}
