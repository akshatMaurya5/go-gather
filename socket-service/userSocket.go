package socketservice

import (
	"encoding/json"
	"go-meta/models"
	"log"
	"math/rand"
	"sync"

	"github.com/gorilla/websocket"
)

var roomManager = GetInstance()

type JWTPayload struct {
	UserId string `json:"userId"`
}

type UserSocket struct {
	models.User
	Mutex sync.Mutex
}

func getRandomString(length int) string {
	characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = characters[rand.Intn(len(characters))]
	}
	return string(result)
}

func newUser(ws *websocket.Conn) *UserSocket {
	return &UserSocket{
		User: models.User{
			ID: getRandomString(10),
			X:  0,
			Y:  0,
			Ws: ws,
		},
	}
}

func (u *UserSocket) handleMessage() {
	defer func() {
		u.Ws.Close()
	}()

	for {
		_, data, err := u.Ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			return
		}
		var message models.Message

		if err := json.Unmarshal(data, &message); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		switch message.Type {
		case "join":
			u.handleJoin(message.Payload)
		case "move":
			u.handleMove(message.Payload)
		default:
			log.Println("Unknown message type:", message.Type)
		}
	}
}

func (u *UserSocket) handleJoin(payload interface{}) {
	data := payload.(map[string]interface{})
	roomId, ok := data["roomId"].(string)
	if !ok {
		u.Ws.Close()
		return
	}

	token, ok := data["token"].(string)
	if !ok {
		u.Ws.Close()
		return
	}

	claims, err := verifyToken(token)
	if err != nil {
		u.Ws.Close()
		return
	}
	u.UserId = claims.UserId
	u.RoomId = roomId

	roomManager.AddUser(roomId, u)

	roomManager.Broadcast(&models.Message{
		Type: "user-joined",
		Payload: map[string]interface{}{
			"userId": u.UserId,
			"x":      u.X,
			"y":      u.Y,
		},
	}, u, roomId)

	u.Send(&models.Message{
		Type: "room-joined",
		Payload: map[string]interface{}{
			"spawn": map[string]interface{}{
				"x": u.X,
				"y": u.Y,
			},
			"users": roomManager.GetUsers(roomId),
		},
	})
}

func (u *UserSocket) handleMove(payload interface{}) {
	data := payload.(map[string]interface{})

	moveX, ok := data["x"].(int)
	if !ok {
		u.Ws.Close()
		return
	}

	moveY, ok := data["y"].(int)
	if !ok {
		u.Ws.Close()
		return
	}

	xDisplacement := abs(moveX - u.X)
	yDisplacement := abs(moveY - u.Y)

	sum := xDisplacement + yDisplacement

	if sum == 1 {
		u.X = moveX
		u.Y = moveY

		roomManager.Broadcast(&models.Message{
			Type: "movement",
			Payload: map[string]interface{}{
				"x": u.X,
				"y": u.Y,
			},
		}, u, u.RoomId)
	} else {
		u.Send(&models.Message{
			Type: "movement-rejected",
			Payload: map[string]interface{}{
				"x": u.X,
				"y": u.Y,
			},
		})
	}
}

func (u *UserSocket) Send(message *models.Message) {
	u.Mutex.Lock()
	defer u.Mutex.Unlock()

	data, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshaling message:", err)
		return
	}

	if err := u.Ws.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Println("Error sending message:", err)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func verifyToken(token string) (*JWTPayload, error) {
	return &JWTPayload{UserId: "dummyUserId"}, nil
}
