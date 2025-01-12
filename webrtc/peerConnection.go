package webrtc

import (
	"log"

	"github.com/pion/webrtc/v4"
)

func (wm *WebRTCManager) CreatePeerConnection(clientID string) (*webrtc.PeerConnection, error) {

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Println("Error creating peer connection:", err)
		return nil, err
	}

	wm.PeerConnections[clientID] = peerConnection

	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			// Handle ICE candidate
			log.Printf("ICE candidate collected for client %s: %v\n", clientID, candidate)
			// Signal candidate to the remote peer via WebSocket
			// Implement the code to send the candidate via the signaling server
		}
	})

	return peerConnection, nil
}

func (wm *WebRTCManager) GetPeerConnection(clientID string) (*webrtc.PeerConnection, bool) {
	pc, exists := wm.PeerConnections[clientID]
	return pc, exists
}

// func (wm *WebRTCManager) ClosePeerConnection(clientID string) {
// 	if pc, exists := wm.PeerConnections[clientID]; exists {
// 		pc.Close()
// 		delete(wm.PeerConnections, clientID)
// 	}
// }
