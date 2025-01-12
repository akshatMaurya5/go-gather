package webrtc

import (
	"encoding/json"
	"fmt"
	"go-gather/types"
	"log"
	"sync"

	"github.com/pion/webrtc/v4"
)

type MessageSender func(clientID string, messageType string, message interface{}) error

type WebRTCManager struct {
	PeerConnections map[string]*webrtc.PeerConnection
	lock            sync.RWMutex
	sendMessage     MessageSender
}

func NewWebRTCManager(sender MessageSender) *WebRTCManager {
	return &WebRTCManager{
		PeerConnections: make(map[string]*webrtc.PeerConnection),
		sendMessage:     sender,
	}
}

func (wm *WebRTCManager) HandleOffer(clientID string, message types.WebRTCMessage) error {
	wm.lock.Lock()
	defer wm.lock.Unlock()

	var sdp types.SDP

	payloadBytes, _ := json.Marshal(message.Payload)

	if err := json.Unmarshal(payloadBytes, &sdp); err != nil {
		log.Println("Error unmarshalling SDP payload:", err)
		return err
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
		return err
	}

	// Store the peer connection
	wm.PeerConnections[clientID] = peerConnection

	// Set the remote description
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp.SDP,
	}

	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Println("Error setting remote description:", err)
		return err
	}

	// Create an answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("Error creating answer:", err)
		return err
	}

	// Set the local description
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Println("Error setting local description:", err)
		return err
	}

	// Prepare the answer to send back via signaling
	answerMessage := types.WebRTCMessage{
		Type:     "webrtc-answer",
		SenderID: clientID,
		TargetID: message.SenderID,
		Payload: types.SDP{
			Type: "answer",
			SDP:  answer.SDP,
		},
	}

	// Use the callback to send the answer
	err = wm.sendMessage(message.SenderID, "webrtc-answer", answerMessage)
	if err != nil {
		log.Println("Error sending answer via callback:", err)
		return err
	}

	// Handle ICE candidates
	peerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}

		candidateInit := candidate.ToJSON()
		iceCandidate := types.ICECandidate{
			Candidate:        candidateInit.Candidate,
			SDPMid:           candidateInit.SDPMid,
			SDPMLineIndex:    candidateInit.SDPMLineIndex,
			UsernameFragment: candidateInit.UsernameFragment,
		}

		candidateMessage := types.WebRTCMessage{
			Type:     "webrtc-candidate",
			SenderID: clientID,
			TargetID: message.SenderID,
			Payload:  iceCandidate,
		}

		// Send the ICE candidate via the callback
		err := wm.sendMessage(message.SenderID, "webrtc-candidate", candidateMessage)
		if err != nil {
			log.Println("Error sending ICE candidate via callback:", err)
		}
	})

	return nil
}

func (wm *WebRTCManager) HandleAnswer(clientID string, message types.WebRTCMessage) error {
	wm.lock.Lock()
	defer wm.lock.Unlock()

	var sdp types.SDP

	payloadBytes, _ := json.Marshal(message.Payload)

	if err := json.Unmarshal(payloadBytes, &sdp); err != nil {
		log.Println("Error unmarshalling SDP payload:", err)
		return err
	}

	peerConnection, exists := wm.PeerConnections[clientID]
	if !exists {
		return fmt.Errorf("peer connection not found for client %s", clientID)
	}

	answer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  sdp.SDP,
	}

	err := peerConnection.SetRemoteDescription(answer)
	if err != nil {
		log.Println("Error setting remote description:", err)
		return err
	}

	return nil
}

func (wm *WebRTCManager) HandleICECandidate(clientID string, message types.WebRTCMessage) error {
	wm.lock.Lock()
	defer wm.lock.Unlock()

	var candidate types.ICECandidate

	payloadBytes, _ := json.Marshal(message.Payload)

	if err := json.Unmarshal(payloadBytes, &candidate); err != nil {
		log.Println("Error unmarshalling ICE candidate payload:", err)
		return err
	}

	peerConnection, exists := wm.PeerConnections[clientID]
	if !exists {
		return fmt.Errorf("peer connection not found for client %s", clientID)
	}

	iceCandidate := webrtc.ICECandidateInit{
		Candidate:        candidate.Candidate,
		SDPMid:           candidate.SDPMid,
		SDPMLineIndex:    candidate.SDPMLineIndex,
		UsernameFragment: candidate.UsernameFragment,
	}

	err := peerConnection.AddICECandidate(iceCandidate)
	if err != nil {
		log.Println("Error adding ICE candidate:", err)
		return err
	}

	return nil
}

func (wm *WebRTCManager) ClosePeerConnection(clientID string) {
	wm.lock.Lock()
	defer wm.lock.Unlock()

	if pc, exists := wm.PeerConnections[clientID]; exists {
		pc.Close()
		delete(wm.PeerConnections, clientID)
	}
}
