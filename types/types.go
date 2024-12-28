package types

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type JoinPayload struct {
	SpaceID string `json:"spaceId"`
	Token   string `json:"token"`
}

type MovePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type User struct {
	ID      string
	UserID  string
	SpaceID string
	X       int
	Y       int
	Conn    *websocket.Conn
}

type Room struct {
	ID    string
	Users map[string]*User
}

var RoomManager = struct {
	Rooms map[string]*Room
}{
	Rooms: make(map[string]*Room),
}

type WebSocketHandler struct {
	Upgrader websocket.Upgrader
}
