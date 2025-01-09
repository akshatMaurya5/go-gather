package main

import (
	"sync"
)

type RTCManager struct {
	rooms map[string][]string // room id -> user id
	lock  sync.RWMutex
}

func NewRTCManager() *RTCManager {
	return &RTCManager{
		rooms: make(map[string][]string),
	}
}

func (rtcM *RTCManager) AddUserToRoom(roomID, userID string) {
	rtcM.lock.Lock()
	defer rtcM.lock.Unlock()
	rtcM.rooms[roomID] = append(rtcM.rooms[roomID], userID)
}

func (rtcM *RTCManager) RemoveUserFromRoom(roomID, userID string) {
	rtcM.lock.Lock()
	defer rtcM.lock.Unlock()

	users, exists := rtcM.rooms[roomID]
	if !exists {
		return
	}

	for i, user := range users {
		if user == userID {
			rtcM.rooms[roomID] = append(users[:i], users[i+1:]...)
			break
		}
	}

	if len(rtcM.rooms[roomID]) == 0 {
		delete(rtcM.rooms, roomID)
	}
}

func (rtcM *RTCManager) GetUsersInRoom(roomID string) []string {
	rtcM.lock.RLock()
	defer rtcM.lock.RUnlock()
	return rtcM.rooms[roomID]
}
