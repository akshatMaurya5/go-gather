package models

type Room struct {
	ID        string
	Clients   map[string]*Socket
	Broadcast chan []byte
	Join      chan []*Socket
	Leave     chan []*Socket
	Move      chan UserMovement
}

type UserMovement struct {
	UserId string `json:"userId"`
	RoomId string `json:"roomId"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}
