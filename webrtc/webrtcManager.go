package webrtc

import (
	"encoding/json"
	"go-gather/types"
	"log"

	"github.com/pion/webrtc/v4"
)

type WebRTCManager struct {
	PeerConnections map[string]*webrtc.PeerConnection
}

func NewWebRTCManager() *WebRTCManager {
	return &WebRTCManager{
		PeerConnections: make(map[string]*webrtc.PeerConnection),
	}
}

func (wm *WebRTCManager) HandleOffer(clientID string, message types.WebRTCMessage) {
	var sdp types.SDP

	payloadBytes, _ := json.Marshal(message.Payload)

	if err := json.Unmarshal(payloadBytes, &sdp); err != nil {
		log.Println("Error unmarshalling SDPs payload:", err)
		return
	}

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
		return
	}

	// LOCALLY STORING THE PEER CONNECTION
	wm.PeerConnections[clientID] = peerConnection

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp.SDP,
	}

	err := peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Println("Error setting remote description & Offer:", err)
		return
	}

	answer, err := peerConnection.CreateAnswer(nil)
}
