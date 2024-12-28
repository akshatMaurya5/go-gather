package socketService

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strings"

	"go-meta/types"

	"github.com/gorilla/websocket"
)

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
	var currentUser *types.User

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("error %s when reading message", err)
			break
		}

		var wsMsg types.WebSocketMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("error unmarshalling message: %v", err)
			break
		}

		switch wsMsg.Type {
		case "join":
			log.Print("Processing join event")
			var payload types.JoinPayload
			if err := json.Unmarshal(wsMsg.Payload, &payload); err != nil {
				log.Printf("error unmarshalling join payload: %v", err)
				break
			}
			handleJoin(c, payload, &currentUser)

		case "move":
			log.Print("Processing move event")
			if currentUser == nil {
				log.Print("Error: No current user found")
				sendMessage(c, map[string]interface{}{
					"type": "error",
					"payload": map[string]string{
						"message": "User not authenticated",
					},
				})
				break
			}
			var payload types.MovePayload
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

func handleJoin(c *websocket.Conn, payload types.JoinPayload, user **types.User) {
	log.Printf("Handling join for space: %s", payload.SpaceID)

	// TODO :  Authenticate user with token
	userID := authenticateUser(payload.Token)

	log.Printf("Authenticated user with ID: %s", userID)
	if userID == "" {
		c.WriteMessage(websocket.CloseMessage, []byte{})
		return
	}

	spaceID := payload.SpaceID
	space, exists := types.RoomManager.Rooms[spaceID]

	if !exists {
		space = &types.Room{ID: spaceID, Users: make(map[string]*types.User)}
		types.RoomManager.Rooms[spaceID] = space
	}
	newUser := &types.User{
		ID:      getRandomString(10),
		UserID:  userID,
		SpaceID: spaceID,
		X:       20, // Updated spawn position
		Y:       20, // Updated spawn position
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

	log.Printf("User spawn position: %d, %d", newUser.X, newUser.Y)

	sendMessage(c, response)

	broadcastUserJoin(space, newUser)
}

func handleMove(c *websocket.Conn, user *types.User, payload types.MovePayload) {
	if user == nil {
		log.Print("Error: User is nil in handleMove")
		return
	}

	xDisplacement := abs(user.X - payload.X)
	yDisplacement := abs(user.Y - payload.Y)
	if (xDisplacement == 1 && yDisplacement == 0) || (xDisplacement == 0 && yDisplacement == 1) {
		user.X = payload.X
		user.Y = payload.Y
		broadcastMovement(user)
		log.Printf("User moved to: %d, %d", user.X, user.Y)
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
	return rand.Intn(62)
}

func getUsersInSpace(space *types.Room, excludeUser *types.User) []map[string]interface{} {
	var users []map[string]interface{}
	for _, u := range space.Users {
		if u.ID != excludeUser.ID {
			users = append(users, map[string]interface{}{
				"id": u.ID,
				"x":  u.X,
				"y":  u.Y,
			})
		}
	}
	return users
}

func sendMessage(c *websocket.Conn, message map[string]interface{}) {
	msg, _ := json.Marshal(message)
	c.WriteMessage(websocket.TextMessage, msg)
}

func broadcastUserJoin(space *types.Room, user *types.User) {
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

	for _, u := range space.Users {
		if u.ID != user.ID {
			sendMessage(user.Conn, map[string]interface{}{
				"type": "user-joined",
				"payload": map[string]interface{}{
					"userId": u.UserID,
					"x":      u.X,
					"y":      u.Y,
				},
			})
		}
	}
}

func broadcastMovement(user *types.User) {
	for _, u := range types.RoomManager.Rooms[user.SpaceID].Users {
		if u.ID != user.ID {
			sendMessage(u.Conn, map[string]interface{}{
				"type": "movement",
				"payload": map[string]interface{}{
					"userId": user.UserID,
					"x":      user.X,
					"y":      user.Y,
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
