package types

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
}

type WebSocketManager struct {
	clients map[string]*Client
	lock    sync.RWMutex
}
