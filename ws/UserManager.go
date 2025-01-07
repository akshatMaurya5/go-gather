package ws

import (
	"encoding/json"
	"fmt"
	"go-gather/types"
	"log"
	"math"
	"net/http"
	"strconv"

	// "go-gather/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()
	userID := r.URL.Query().Get("userId")
	roomID := r.URL.Query().Get("roomId")

	if userID == "" || roomID == "" {
		log.Println("userId or roomId is missing")
		conn.WriteMessage(websocket.TextMessage, []byte("userId and roomId are required"))
		return
	}

	// Get WebSocket manager singleton
	wsManager := GetWebSocketInstance()

	client := &Client{
		ID:     userID,
		roomID: roomID,
		Conn:   conn,
		X:      0,
		Y:      0,
	}
	wsManager.AddUser(client, roomID)

	for {
		_, messageBytes, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading types.Message from client %s: %v\n", userID, err)
			wsManager.RemoveUser(userID, roomID)
			break
		}

		var message types.Message
		if err := json.Unmarshal(messageBytes, &message); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		handleEvents(wsManager, client, roomID, message)
	}
}

func handleEvents(wsManager *WebSocketManager, client *Client, roomID string, message types.Message) {
	var response types.Response

	switch message.Type {
	case "join":
		success := handleJoinRoom(wsManager, client, roomID)
		response = types.Response{
			Type:    "user-joined",
			Success: success,
			Data: map[string]string{
				"userId": client.ID,
				"roomId": roomID,
				"X":      strconv.Itoa(client.X),
				"Y":      strconv.Itoa(client.Y),
			},
		}
	case "leave-room":
		success := handleLeaveRoom(wsManager, client, roomID)
		response = types.Response{
			Type:    "user-left",
			Success: success,
			Data: map[string]string{
				"userId": client.ID,
				"roomId": roomID,
			},
		}
	case "send-types.Message":
		messageStr, ok := message.Data.(string)
		if !ok {
			response = types.Response{
				Type:    "error",
				Success: false,
				Error:   "Invalid types.Message format",
			}
		} else {
			wsManager.BroadcastToRoom(roomID, messageStr)
			response = types.Response{
				Type:    "message-sent",
				Success: true,
				Data: map[string]string{
					"message": messageStr,
				},
			}
		}
	case "move":
		var moveData types.MoveData
		dataBytes, _ := json.Marshal(message.Data)
		if err := json.Unmarshal(dataBytes, &moveData); err != nil {
			response = types.Response{
				Type:    "error",
				Success: false,
				Error:   "Invalid move data",
			}
		} else {
			success := handleMove(wsManager, client, roomID, moveData)
			response = types.Response{
				Type:    "move-completed",
				Success: success,
				Data: map[string]int{
					"x": client.X,
					"y": client.Y,
				},
			}
		}
	default:
		response = types.Response{
			Type:    "error",
			Success: false,
			Error:   "Unknown event type",
		}
	}

	client.SendMessage(response.Type, response)
}

func handleJoinRoom(wsManager *WebSocketManager, client *Client, roomID string) bool {

	wsManager.AddUser(client, roomID)
	log.Printf("User %s joined room %s\n", client.ID, roomID)
	wsManager.BroadcastToRoom(roomID, fmt.Sprintf("%s joined the room at %d,%d", client.ID, client.X, client.Y))
	return true
}

func handleLeaveRoom(wsManager *WebSocketManager, client *Client, roomID string) bool {
	log.Printf("User %s left room %s\n", client.ID, roomID)
	wsManager.RemoveUser(client.ID, roomID)
	wsManager.BroadcastToRoom(roomID, fmt.Sprintf("%s left the room", client.ID))
	return true
}

func handleMove(wsManager *WebSocketManager, client *Client, roomID string, moveData types.MoveData) bool {
	log.Printf("User %s moved in room %s\n", client.ID, roomID)

	xDisplacement := math.Abs(float64(client.X - moveData.X))
	yDisplacement := math.Abs(float64(client.Y - moveData.Y))

	if xDisplacement+yDisplacement > 1 {
		log.Println("Invalid move")
		return false
	}

	client.X = moveData.X
	client.Y = moveData.Y
	wsManager.BroadcastMove(client, roomID)
	return true
}
