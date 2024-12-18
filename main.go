package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// Initialize the random seed
func init() {
	rand.Seed(time.Now().UnixNano())
}

// WebSocket message structure
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Define the payload for join and move actions
type JoinPayload struct {
	SpaceID string `json:"spaceId"`
	Token   string `json:"token"`
}

type MovePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// User structure (equivalent to User class in JS)
type User struct {
	ID      string
	UserID  string
	SpaceID string
	X       int
	Y       int
	Conn    *websocket.Conn
}

// Room structure to store users
type Room struct {
	ID    string
	Users map[string]*User
}

// Room manager (equivalent to RoomManager in JS)
var RoomManager = struct {
	Rooms map[string]*Room
}{
	Rooms: make(map[string]*Room),
}

// WebSocket handler
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

		log.Printf("Received message: %s", msg)

		// Parse the message to determine the type
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

// Handle the "join" event
func handleJoin(c *websocket.Conn, payload JoinPayload, user **User) {
	log.Printf("Handling join for space: %s", payload.SpaceID)
	// Authenticate user with token (simplified)
	userID := authenticateUser(payload.Token)

	log.Printf("Authenticated user with ID: %s", userID)
	if userID == "" {
		c.WriteMessage(websocket.CloseMessage, []byte{})
		return
	}

	// Create or join a space
	spaceID := payload.SpaceID
	space, exists := RoomManager.Rooms[spaceID]
	if !exists {
		space = &Room{ID: spaceID, Users: make(map[string]*User)}
		RoomManager.Rooms[spaceID] = space
	}

	// Create new user and add to room
	newUser := &User{
		ID:      getRandomString(10),
		UserID:  userID,
		SpaceID: spaceID,
		X:       0, // Spawn at a default position
		Y:       0, // Default Y position
		Conn:    c,
	}
	space.Users[newUser.ID] = newUser
	*user = newUser

	log.Printf("Current users in space %s: %v", spaceID, space.Users)

	// Send response back to client
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

	// Broadcast user join to other users in the room
	broadcastUserJoin(space, newUser)
}

// Handle the "move" event
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

// Helper functions
func authenticateUser(token string) string {
	// Implement token verification and return a random string of length 5 if valid (simplified here)
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
	return rand.Intn(10) // 62 is the length of the characters string
}

func getUsersInSpace(space *Room, excludeUser *User) []map[string]interface{} {
	var users []map[string]interface{}
	for _, u := range space.Users {
		// Always include the current user if they are the only user
		// if len(space.Users) == 1 || u.ID != excludeUser.ID {
		users = append(users, map[string]interface{}{
			"id": u.ID,
		})
		// }
	}
	log.Printf("Users in space %s (excluding %s): %v", space.ID, excludeUser.ID, users)
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

func main() {
	log.Print("Starting server...")
	wsh := WebSocketHandler{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	http.Handle("/", wsh)
	log.Print("WebSocket handler registered. Listening on port 3001...")
	log.Fatal(http.ListenAndServe(":3001", nil))
	// log.Print("WebSocket server started on port 3001")
}
