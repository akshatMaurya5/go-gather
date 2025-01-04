package types

import "github.com/gorilla/websocket"

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Response struct {
	Type    string      `json:"type"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type MoveData struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Client struct {
	ID     string
	roomID string
	Conn   *websocket.Conn
	X      int
	Y      int
}

type Room struct {
	ID      string
	clients map[string]*Client
}

type Client struct {
	ID     string
	roomID string
	Conn   *websocket.Conn
	X      int
	Y      int
}
