﻿# golang - meta
https://v0.dev/chat/GeMKjDiBNLD



### WEBRTC LEARNINGS:

#### SIGNALING SERVER: (has ws integration)
- **Offer**: 
    - First message sent 
    - Contains SDP (Session Description Protocol) which contains information about media types/codecs and other settings

- **Answer**:
    - Receiving client accepts or rejects the offer 
    - Contains SDP information that matches the offer

- **ICE Candidates**:
    - Contains information about the network interfaces and public IP addresses of the client

#### PEER CONNECTION MANAGEMENT:
    - Handles audio/video streams

#### RTC MANAGER:
    - Creates a central point to manage signaling, room management, and peer connections

### Sequence Diagram
![image](https://github.com/user-attachments/assets/06affa03-55e6-468f-b1ea-0109c2ece5cb)


