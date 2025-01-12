package ws

import (
	"encoding/json"
	"fmt"
	"go-gather/types"
	"go-gather/webrtc"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Function to send messages to clients
func sendMessageToClient(clientID string, messageType string, message interface{}) error {
	client := wsManager.GetClientByID(clientID)
	if client == nil {
		return fmt.Errorf("client %s not found", clientID)
	}
	client.SendMessage(messageType, message)
	return nil
}

// Instantiate the WebRTCManager with the sendMessage function
var webrtcManager = webrtc.NewWebRTCManager(sendMessageToClient)

// Instantiate the WebSocketManager singleton
var wsManager = GetWebSocketInstance()

func HandleWebsocket(w http.ResponseWriter, r *http.Request) {
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
			log.Printf("Error reading Message from client %s: %v\n", userID, err)
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

	log.Printf("Received message type: %s from user: %s\n", message.Type, client.ID)

	switch message.Type {
	case "join":
		success := handleJoinRoom(wsManager, client, roomID)

		if success {
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
		} else {
			response = types.Response{
				Type:    "user-joining-failed",
				Success: false,
				Error:   "User does not have access to this room",
			}
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

	case "send-message":
		messageStr, ok := message.Data.(string)
		if !ok {
			response = types.Response{
				Type:    "error",
				Success: false,
				Error:   "Invalid message format",
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

	case "webrtc-offer", "webrtc-answer", "webrtc-candidate":
		handleWebRTCSignaling(wsManager, client, message)
		// No response needed for signaling messages
		return

	default:
		response = types.Response{
			Type:    "error",
			Success: false,
			Error:   "Unknown event type",
		}
	}

	// Send response if it's set
	if response.Type != "" {
		client.SendMessage(response.Type, response)
	}
}

func handleWebRTCSignaling(wsManager *WebSocketManager, client *Client, message types.Message) {
	var webrtcMessage types.WebRTCMessage
	dataBytes, _ := json.Marshal(message.Data)
	if err := json.Unmarshal(dataBytes, &webrtcMessage); err != nil {
		log.Println("Invalid WebRTC message format:", err)
		return
	}

	switch message.Type {
	case "webrtc-offer":
		err := webrtcManager.HandleOffer(client.ID, webrtcMessage)
		if err != nil {
			log.Println("Error handling offer:", err)
		}
	case "webrtc-answer":
		err := webrtcManager.HandleAnswer(client.ID, webrtcMessage)
		if err != nil {
			log.Println("Error handling answer:", err)
		}
	case "webrtc-candidate":
		err := webrtcManager.HandleICECandidate(client.ID, webrtcMessage)
		if err != nil {
			log.Println("Error handling ICE candidate:", err)
		}
	default:
		log.Println("Unknown WebRTC message type:", message.Type)
	}
}

func handleJoinRoom(wsManager *WebSocketManager, client *Client, roomID string) bool {
	log.Println("handleJoinRoom called")
	url := fmt.Sprintf("http://localhost:3000/authenticate?email=%s", client.ID)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error calling authenticate endpoint: %v", err)
		return false
	}
	defer resp.Body.Close()

	var result struct {
		EmailAddress string   `json:"emailAddress"`
		Rooms        []string `json:"rooms"`
		Success      bool     `json:"success"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %v", err)
		return false
	}

	if !result.Success {
		log.Printf("Error: API returned unsuccessful response")
		return false
	}

	log.Println("These are the rooms:", result.Rooms)

	rooms := result.Rooms

	isAuthorized := false

	for _, room := range rooms {
		if room == roomID {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		return false
	}

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
