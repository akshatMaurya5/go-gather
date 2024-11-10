package models

import "github.com/gorilla/websocket"

type Socket struct {
	Conn   *websocket.Conn
	Send   chan []byte
	UserId string
	RoomId string
}
