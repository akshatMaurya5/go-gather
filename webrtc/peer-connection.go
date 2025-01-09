package main

import (
	"log"

	"github.com/pion/webrtc/v3"
)

var (
	peerConnManager = NewPeerConnectionManager()
	rtcManager      = NewRTCManager()
	wsManager       = GetWebSocketInstance()
)

func NewPeerConnectionManager() *PeerConnectionManager {
	return &PeerConnectionManager{
		peerConnections: make(map[string]*webrtc.PeerConnection),
	}
}

func (pcm *PeerConnectionManager) CreatePeerConnection(userID string) (*webrtc.PeerConnection, error) {

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

	pcm.peerConnections[userID] = peerConnection
	return peerConnection, nil
}
