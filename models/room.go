package models

type Room struct {
	ID         string
	Clients    map[string]*Socket
	Broadcasts chan []byte
	Join       chan []*Socket
	Leave      chan []*Socket
}
