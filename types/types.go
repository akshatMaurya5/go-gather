package types

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type Response struct {
	Type    string      `json:"type"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type MoveData struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WebRTC-related types
type WebRTCMessage struct {
	Type     string      `json:"type"`     // "webrtc-offer", "webrtc-answer", "webrtc-candidate"
	SenderID string      `json:"senderId"` // Who sent the message
	TargetID string      `json:"targetId"` // Who should receive the message
	Payload  interface{} `json:"payload"`  // SDP or ICE Candidate
}

type SDP struct {
	Type string `json:"type"` // "offer" or "answer"
	SDP  string `json:"sdp"`
}

type ICECandidate struct {
	Candidate        string  `json:"candidate"`
	SDPMid           *string `json:"sdpMid,omitempty"`
	SDPMLineIndex    *uint16 `json:"sdpMLineIndex,omitempty"`
	UsernameFragment *string `json:"usernameFragment,omitempty"`
}
