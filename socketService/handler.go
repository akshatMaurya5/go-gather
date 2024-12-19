package socketService

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type JoinPayload struct {
	SpaceID string `json:"spaceId"`
	Token   string `json:"token"`
}

type MovePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type User struct {
	ID      string
	UserID  string
	SpaceID string
	X       int
	Y       int
	Conn    *websocket.Conn
}

type Room struct {
	ID    string
	Users map[string]*User
}

var RoomManager = struct {
	Rooms map[string]*Room
}{
	Rooms: make(map[string]*Room),
}

type WebSocketHandler struct {
	Upgrader websocket.Upgrader
}

func (wsh WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Print("New WebSocket connection attempt")
	c, err := wsh.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error %s when upgrading connection to websocket", err)
		return
	}
	defer c.Close()

	log.Print("Client connected")
	var currentUser *User

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("error %s when reading message", err)
			break
		}

		var wsMsg WebSocketMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("error unmarshalling message: %v", err)
			break
		}

		switch wsMsg.Type {
		case "join":
			log.Print("Processing join event")
			var payload JoinPayload
			if err := json.Unmarshal(wsMsg.Payload, &payload); err != nil {
				log.Printf("error unmarshalling join payload: %v", err)
				break
			}
			handleJoin(c, payload, &currentUser)

		case "move":
			log.Print("Processing move event")
			var payload MovePayload
			if err := json.Unmarshal(wsMsg.Payload, &payload); err != nil {
				log.Printf("error unmarshalling move payload: %v", err)
				break
			}
			handleMove(c, currentUser, payload)

		default:
			log.Printf("unknown event type: %s", wsMsg.Type)
		}
	}
}

func handleJoin(c *websocket.Conn, payload JoinPayload, user **User) {
	log.Printf("Handling join for space: %s", payload.SpaceID)

	// TODO :  Authenticate user with token
	userID := authenticateUser(payload.Token)

	log.Printf("Authenticated user with ID: %s", userID)
	if userID == "" {
		c.WriteMessage(websocket.CloseMessage, []byte{})
		return
	}

	spaceID := payload.SpaceID
	space, exists := RoomManager.Rooms[spaceID]

	//  Create a new space if it doesn't exist
	if !exists {
		space = &Room{ID: spaceID, Users: make(map[string]*User)}
		RoomManager.Rooms[spaceID] = space
	}

	// Create new user and add to room
	newUser := &User{
		ID:      getRandomString(10),
		UserID:  userID,
		SpaceID: spaceID,
		X:       0,
		Y:       0,
		Conn:    c,
	}
	space.Users[newUser.ID] = newUser
	*user = newUser

	log.Printf("Current users in space %s: %v", spaceID, space.Users)

	response := map[string]interface{}{
		"type": "space-joined",
		"payload": map[string]interface{}{
			"spawn": map[string]int{
				"x": newUser.X,
				"y": newUser.Y,
			},
			"users": getUsersInSpace(space, newUser),
		},
	}
	sendMessage(c, response)

	broadcastUserJoin(space, newUser)
}

func handleMove(c *websocket.Conn, user *User, payload MovePayload) {
	// Simple validation (only allow moves in adjacent tiles)
	xDisplacement := abs(user.X - payload.X)
	yDisplacement := abs(user.Y - payload.Y)
	if (xDisplacement == 1 && yDisplacement == 0) || (xDisplacement == 0 && yDisplacement == 1) {
		user.X = payload.X
		user.Y = payload.Y
		broadcastMovement(user)
	} else {
		sendMessage(c, map[string]interface{}{
			"type": "movement-rejected",
			"payload": map[string]int{
				"x": user.X,
				"y": user.Y,
			},
		})
	}
}

func authenticateUser(token string) string {
	return getRandomString(5)
}
func getRandomString(length int) string {
	const characters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	var result strings.Builder
	for i := 0; i < length; i++ {
		result.WriteByte(characters[int(randomInt())])
	}
	return result.String()
}

func randomInt() int {
	return rand.Intn(10)
}

func getUsersInSpace(space *Room, excludeUser *User) []map[string]interface{} {
	var users []map[string]interface{}
	for _, u := range space.Users {
		users = append(users, map[string]interface{}{
			"id": u.ID,
		})
	}
	return users
}

func sendMessage(c *websocket.Conn, message map[string]interface{}) {
	msg, _ := json.Marshal(message)
	c.WriteMessage(websocket.TextMessage, msg)
}

func broadcastUserJoin(space *Room, user *User) {
	for _, u := range space.Users {
		if u.ID != user.ID {
			sendMessage(u.Conn, map[string]interface{}{
				"type": "user-joined",
				"payload": map[string]interface{}{
					"userId": user.UserID,
					"x":      user.X,
					"y":      user.Y,
				},
			})
		}
	}
}

func broadcastMovement(user *User) {
	for _, u := range RoomManager.Rooms[user.SpaceID].Users {
		if u.ID != user.ID {
			sendMessage(u.Conn, map[string]interface{}{
				"type": "movement",
				"payload": map[string]int{
					"x": user.X,
					"y": user.Y,
				},
			})
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
