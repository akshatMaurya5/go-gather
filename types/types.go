package types

import "encoding/json"

type webSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type JoinPayload struct {
	SpaceId string `json:"spaceId"`
	Token   string `json:"token"`
}

type User struct {
	Id      string `json:"id"`
	UserId  string `json:"userId"`
	SpaceId string `json:"spaceId"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
}

type Room struct {
	ID    string `json:"id"`
	Users map[string]*User
}
