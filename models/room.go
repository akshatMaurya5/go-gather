package models

type Room struct {
	ID        string
	Clients   map[string]*Socket
	Broadcast chan []byte
	Join      chan []*Socket
	Leave     chan []*Socket
}
