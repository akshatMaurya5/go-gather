package models

type User struct {
	ID     string `bson:"_id"`
	X      int    `bson:"x"`
	Y      int    `bson:"y"`
	Name   string `bson:"name"`
	RoomID string `bson:"room_id"`
}
