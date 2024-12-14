package models

type Message struct {
	Type    string      `bson:"type"`
	Payload interface{} `bson:"payload"`
}

type JoinPayload struct {
	RoomId string `bson:"roomId"`
	Token  string `bson:"token"`
}

type MovePayload struct {
	X int `bson:"x"`
	Y int `bson:"y"`
}
