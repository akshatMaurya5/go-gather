package main

import (
	"go-gather/ws"
	"log"
	"net/http"
)

func main() {
	// Use ws.HandleWebsocket instead of ws.NewWebSocketHandler
	http.HandleFunc("/ws", ws.HandleWebsocket)

	log.Println("Server started at :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
