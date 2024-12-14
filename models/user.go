package models

import "github.com/gorilla/websocket"

type User struct {
	ID     string `bson:"_id"`
	UserId string `bson:"user_id"`
	RoomId string `bson:"room_id"`
	X      int    `bson:"x"`
	Y      int    `bson:"y"`
	Ws     *websocket.Conn
}
