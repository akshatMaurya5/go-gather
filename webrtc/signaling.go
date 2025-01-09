package main

import (
	"encoding/json"
	"go-gather/types"
	"log"

	"github.com/gorilla/websocket"
)

type SignalingMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func HandleSignaling(conn *websocket.Conn) {
	defer conn.Close()

	for {
		_, incomingMsg, err := conn.ReadMessage()

		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		var msg types.SignalingMessage

		if err := json.Unmarshal(incomingMsg, &msg); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		switch msg.Type {
		case "offer":
			userID := msg.Data.(map[string]interface{})["userId"].(string)
			peerConnection, err := peerConnManager.CreatePeerConnection(userID)
			if err != nil {
				log.Println("Error creating peer connection:", err)
				return
			}
			handleOffer(peerConnection, signalingMessage)
		default:
			log.Println("Unknown message type:", msg.Type)
		}
	}
}

func HandleWebRTCMessage(conn *websocket.Conn, msg types.SignalingMessage) {

	roomId, userID := getRoomAndUserID(conn)

	rtcManager := NewRTCManager()

	switch msg.Type {
	case "offer":
		targetUserID := msg.Data.(map[string]interface{})["targetUserID"].(string)

		forwardMessage(conn, rtcManager, roomID, targetUserID, msg)

	case "answer":
		originUserID := msg.Data.(map[string]interface{})["originUserID"].(string)
		forwardMessage(conn, rtcManager, roomID, targetUserID, msg)

	case "ice-candidate":
		targetUserID := msg.Data.(map[string]interface{})["targetUserID"].(string)
		forwardMessage(conn, rtcManager, roomID, targetUserID, msg)

	default:
		log.Println("Unknown message type:", msg.Type)
	}

}

func forwardMessage(conn *websocket.Conn, rtcManager *RTCManager, roomID, targetUserID string, msg types.SignalingMessage) {
	users := rtcManager.GetUsersInRoom(roomID)
	for _, user := range users {
		if user == targetUserID {
			targetConn := GetWebSocketInstance().GetUserConnection(targetUserID)

			if targetConn != nil {
				targetConn.WriteMessage(websocket.TextMessage, []byte(msg))
			}
		}
	}
}
